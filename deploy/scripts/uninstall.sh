#!/bin/bash
set -e

echo
echo
echo "AWS resources created with Terraform will be destroyed"
pushd $PWD/deploy/terraform
terraform init
terraform destroy
popd
echo
echo
echo "DonkeyVPN resources uninstalled from your AWS account"
