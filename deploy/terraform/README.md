Deploy Donkey VPN Infrastructure
===

This deploys the basic infrastructure used by the DonkeyVPN Telegram Bot deployed using a Lambda Function.

Check the `example.tfvars` `variables.tf` files for the complete list of optional or required inputs for this module.

In order to store correctly the Terraform state file, send the proper configurations for the S3 backed. Use the `example.tfbackend` file as an example.

## Applying changes
Follow the next instructions:

```sh
$ cp example.tfbackend terraform.tfbackend
$ # Update terraform.tfbackend accordingly
$ cp example.tfvars terraform.tfvars
$ # Update terraform.tfvars accordingly
$ terraform init -backend-config terraform.tfbackend
$ terraform plan
$ terraform apply
```
