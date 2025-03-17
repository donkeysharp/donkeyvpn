locals {
  debian_ami_owner = "136693071363"

  default_prefix = "donkeyvpn-${var.environment}"
  prefix         = var.prefix != "" ? var.prefix : local.default_prefix

  ec2_role_name = "${local.prefix}-ec2"

  dynamodb_peers_table_name = "${local.prefix}-peers"
  dynamodb_instances_table_name = "${local.prefix}-instances"

  base_tags = {
    Project     = "donkeyvpn"
    Environment = var.environment
    ManagedBy   = "terraform"
  }

  autoscaling_group_name = local.prefix

  asg_tags = merge({
    Name = local.prefix
  }, local.base_tags)

  account_id = data.aws_caller_identity.current.account_id
  region     = data.aws_region.current.name
}

terraform {
  backend "s3" {
  }
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

data "aws_ami" "debian" {
  most_recent = true
  owners      = [local.debian_ami_owner]

  filter {
    name   = "name"
    values = ["debian-12-amd64*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

data "aws_kms_alias" "ssm" {
  name = var.kms_key_alias
}
