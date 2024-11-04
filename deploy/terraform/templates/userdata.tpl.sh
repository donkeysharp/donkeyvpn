#!/bin/bash

function log() {
    echo "DonkeyVPN - $@"
}

function prepare_dependencies() {
    log "DonkeyPreparing dependencies"

    sudo apt update
    sudo apt install -y wireguard unzip
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

    aws route53 change-resource-record-sets \
        --hosted-zone-id "$HOSTED_ZONE_ID" \
        --change-batch "$CHANGE_BATCH"

    log "Route53 update finished"
}

function configure_wireguard() {
    log "Configuring Wireguard"
}

log "Initializing all the configuration process..."

prepare_dependencies
update_route53
configure_wireguard
