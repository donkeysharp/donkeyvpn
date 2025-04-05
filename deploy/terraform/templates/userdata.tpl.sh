#!/bin/bash
set -e

function log() {
    echo "DonkeyVPN - $@"
}

function notify() {
    status=$1
    log "Sending notificatoin with status: $status"
    BODY="{\"hostname\": \"$DOMAIN_NAME\", \"status\": \"$status\", \"instanceId\": \"$INSTANCE_ID\", \"port\": \"$PORT\" }"
    log "Notification body: $BODY"
    result=$(curl -sS -H 'content-type: application/json' \
        -H "x-api-key: $API_SECRET" \
        --data "$BODY" \
        -XPOST "$API_BASE_URL/v1/api/vpn/notify/$VPN_INSTANCE_ID")
    log "result: $result"
}

function handle_error() {
    log "Handling error, notifying failure"
    notify "error"
}
trap 'handle_error' ERR

function load_settings() {
    export API_BASE_URL=${in_api_base_url}
    export API_SECRET=${in_api_secret}
    export PORT="51820"
    export USE_ROUTE53=${in_use_route53}
    export PUBLIC_IP=$(curl -sS -L ifconfig.me)
    export TOKEN=$(curl -sS -X PUT -H "X-aws-ec2-metadata-token-ttl-seconds: 21600" "http://169.254.169.254/latest/api/token" )
    export REGION=$(curl -sS -H "X-aws-ec2-metadata-token: $TOKEN" http://169.254.169.254/latest/meta-data/placement/region)
    export INSTANCE_ID=$(curl -sS -H "X-aws-ec2-metadata-token: $TOKEN" http://169.254.169.254/latest/meta-data/instance-id)
}

function configure_domain_name() {
    log "Getting next id"
    export VPN_INSTANCE_ID=$(curl -sS -H "x-api-key: $API_SECRET" "$API_BASE_URL/v1/api/vpn/pending" | jq -r '.[0].Id')
    log "VPN_INSTANCE_ID: $VPN_INSTANCE_ID"

    if [[ $USE_ROUTE53 == "true" ]]; then
        export DOMAIN_NAME="$VPN_INSTANCE_ID.${in_domain_name}"
    else
        export DOMAIN_NAME=$PUBLIC_IP
    fi
    log "Instance domain name: $DOMAIN_NAME"
}

function prepare_dependencies() {
    log "DonkeyVPN. Preparing dependencies"

    sudo apt update
    sudo apt install -y wireguard unzip jq
    pushd /opt
        curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
        unzip -q awscliv2.zip
        sudo ./aws/install
    popd

    log "Dependencies ready"
}

function update_route53() {
    if [[ $USE_ROUTE53 != "true" ]]; then
        log "Skipping route53 registration. Disabled."
        return 0
    fi
    log "Updating DNS records"

    HOSTED_ZONE_ID="${in_hosted_zone_id}"
    TTL="${in_vpn_record_ttl}"

    log "HOSTED_ZONE=$HOSTED_ZONE_ID DOMAIN_NAME=$DOMAIN_NAME PUBLIC_IP=$PUBLIC_IP TTL=$TTL"

    # Create JSON payload for UPSERT
    CHANGE_BATCH=$(cat <<EOF
{
    "Comment": "Update A record to public IP",
    "Changes": [
        {
            "Action": "UPSERT",
            "ResourceRecordSet": {
                "Name": "$DOMAIN_NAME",
                "Type": "A",
                "TTL": $TTL,
                "ResourceRecords": [
                    {
                        "Value": "$PUBLIC_IP"
                    }
                ]
            }
        }
    ]
}
EOF
)
    log "Updating Route53 record..."
    aws route53 change-resource-record-sets \
        --hosted-zone-id "$HOSTED_ZONE_ID" \
        --change-batch "$CHANGE_BATCH"

    log "Route53 update finished"
}

function configure_peers() {
    log "Retrieving list of peers from API"
    curl -sS -H "x-api-key: $API_SECRET" "$API_BASE_URL/v1/api/peer" -o /tmp/wireguard_peers.txt

    log "Adding peers to wg0.conf"
    for peer in $(cat /tmp/wireguard_peers.txt); do
        PEER_IP_ADDRESS=$(echo $peer | cut -d',' -f1)
        PEER_PUBLIC_KEY=$(echo $peer | cut -d',' -f2)
        log "Adding peer: $PEER_IP_ADDRESS"
        cat <<EOF >> /etc/wireguard/wg0.conf
[Peer]
PublicKey = $PEER_PUBLIC_KEY
AllowedIPs = $PEER_IP_ADDRESS/32

EOF
    done
}

function configure_wireguard() {
    log "Configuring Wireguard"

    log "Enabling ip_forward"
    echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
    sysctl -p

    PRIVATE_KEY_SSM_PARAM=${in_ssm_private_key}
    PUBLIC_KEY_SSM_PARAM=${in_ssm_public_key}

    log "Setting up private and public keys"
    log "PRIVATE_KEY_SSM_PARAM=$PRIVATE_KEY_SSM_PARAM"
    log "PUBLIC_KEY_SSM_PARAM=$PUBLIC_KEY_SSM_PARAM"
    PRIVATE_KEY=$(aws ssm get-parameter \
        --name "$PRIVATE_KEY_SSM_PARAM" \
        --with-decryption \
        --region $REGION \
        --query "Parameter.Value" \
        --output text)

    PUBLIC_KEY=$(aws ssm get-parameter \
        --name "$PUBLIC_KEY_SSM_PARAM" \
        --with-decryption \
        --region $REGION \
        --query "Parameter.Value" \
        --output text)

    echo $PRIVATE_KEY > /etc/wireguard/privatekey
    echo $PUBLIC_KEY > /etc/wireguard/publickey

    cat <<EOF >> /etc/wireguard/wg0.conf
[Interface]
Address = ${in_wg_interface_address}
ListenPort = $PORT
PrivateKey = $PRIVATE_KEY

EOF
    log "Wireguard configuration finished"

    configure_peers

    chmod 700 /etc/wireguard/*

    log "Route all internet traffic via the VPN"
    iptables -t nat -A POSTROUTING -s ${in_wg_ip_range} -o ens5 -j MASQUERADE

    wg-quick up wg0
    systemctl enable wg-quick@wg0
}

log "Initializing all the configuration process..."

prepare_dependencies
load_settings
configure_domain_name
configure_wireguard
update_route53
notify "success"
