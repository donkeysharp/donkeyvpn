#!/bin/bash

function log() {
    echo "DonkeyVPN - $@"
}

function prepare_dependencies() {
    log "DonkeyPreparing dependencies"

    sudo apt update
    sudo apt install -y wireguard unzip jq
    pushd /opt
        curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
        unzip awscliv2.zip
        sudo ./aws/install
    popd

    log "Dependencies ready"
}

function update_route53() {
    log "Updating DNS records"

    HOSTED_ZONE_ID="${in_hosted_zone_id}"
    DOMAIN_NAME="${in_vpn_record_name}"
    PUBLIC_IP=$(curl -sS -L ifconfig.me)
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

function configure_wireguard() {
    log "Configuring Wireguard"

    export TOKEN=$(curl -sS -X PUT -H "X-aws-ec2-metadata-token-ttl-seconds: 21600" "http://169.254.169.254/latest/api/token" )
    export REGION=$(curl -sS -H "X-aws-ec2-metadata-token: $TOKEN" http://169.254.169.254/latest/meta-data/placement/region)

    PRIVATE_KEY_SSM_PARAM=${in_ssm_private_key}
    PUBLIC_KEY_SSM_PARAM=${in_ssm_public_key}
    PEERS_SSM_PARAM=${in_ssm_peers}

    log "Setting up private and public keys"
    log "PRIVATE_KEY_SSM_PARAM=$PRIVATE_KEY_SSM_PARAM"
    log "PUBLIC_KEY_SSM_PARAM=$PUBLIC_KEY_SSM_PARAM"
    PRIVATE_KEY=$(aws ssm get-parameter \
        --name "/$PRIVATE_KEY_SSM_PARAM" \
        --with-decryption \
        --region $REGION \
        --query "Parameter.Value" \
        --output text)

    PUBLIC_KEY=$(aws ssm get-parameter \
        --name "/$PUBLIC_KEY_SSM_PARAM" \
        --with-decryption \
        --region $REGION \
        --query "Parameter.Value" \
        --output text)

    aws ssm get-parameter \
        --name "/$PEERS_SSM_PARAM" \
        --with-decryption \
        --region $REGION \
        --query "Parameter.Value" \
        --output text >> /tmp/wireguard_peers.txt

    echo $PRIVATE_KEY > /etc/wireguard/privatekey
    echo $PUBLIC_KEY > /etc/wireguard/publickey

    log "Enabling ip_forward"
    echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
    sysctl -p

    log "Wireguard configuration finished"
    cat <<EOF >> /etc/wireguard/wg0.conf
[Interface]
Address = 10.0.0.1/24
ListenPort = 51820
PrivateKey = $PRIVATE_KEY

EOF
    log "Adding peers to wg0.conf"
    for peer in $(cat /tmp/wireguard_peers.txt); do
        PEER_IP_ADDRESS=$(echo $peer | cut -d',' -f1)
        PEER_PUBLIC_KEY=$(echo $peer | cut -d',' -f2)
        echo "Adding peer: $PEER_IP_ADDRESS"
        cat <<EOF >> /etc/wireguard/wg0.conf
[Peer]
PublicKey = $PEER_PUBLIC_KEY
AllowedIPs = $PEER_IP_ADDRESS/32

EOF
    done

    chmod 700 /etc/wireguard/*

    log "Route all internet traffic via the VPN"
    iptables -t nat -A POSTROUTING -s 10.0.0.0/24 -o ens5 -j MASQUERADE

    wg-quick up wg0
    systemctl enable wg-quick@wg0
}

log "Initializing all the configuration process..."

prepare_dependencies
update_route53
configure_wireguard
