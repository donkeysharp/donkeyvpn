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

  # In case var.testing_userdata_api_base_url is not set, it will use API Gateway's url
  api_base_url = var.testing_userdata_api_base_url == "" ? trim(aws_apigatewayv2_stage.default.invoke_url, "/") : var.testing_userdata_api_base_url
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

data "aws_ssm_parameter" "telegram_bot_api_token" {
  name            = var.telegram_bot_api_token_ssm_param
  with_decryption = true
}

data "aws_ssm_parameter" "webhook_secret" {
  name            = var.webhook_secret_ssm_param
  with_decryption = true
}

resource "local_file" "webhook_register" {
  content  = templatefile("${path.module}/templates/webhook-register.tpl.sh", {
    IN_TELEGRAM_BOT_API_TOKEN = data.aws_ssm_parameter.telegram_bot_api_token.value
    IN_SECRET_TOKEN           = data.aws_ssm_parameter.webhook_secret.value
    IN_BASE_URL               = trim(aws_apigatewayv2_stage.default.invoke_url, "/")
  })
  file_permission = "0755"
  filename = "/tmp/donkeyvpn-webhook-register.sh"
}
