#!/bin/bash
set -e

go run cmd/installer/main.go \
    -tfvars-template "$PWD/deploy/terraform/terraform.tfvars.example" \
    -tfvars-output "$PWD/deploy/terraform/terraform.tfvars" \
    -tfbackend-template "$PWD/deploy/terraform/terraform.tfbackend.example" \
    -tfbackend-output "$PWD/deploy/terraform/terraform.tfbackend"

echo
echo
echo "Building Donkeyvpn binary..."
go build -o dist/bootstrap cmd/bot/main.go
echo "DonkeyVPN built successfully"

echo
echo
echo "Starting Terraform plan and apply"
pushd $PWD/deploy/terraform
terraform init
terraform apply
popd
echo
echo
echo "Installation finished successfully! Send /help to your telegram bot"
