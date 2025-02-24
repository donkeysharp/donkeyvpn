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

resource "aws_security_group" "default" {
  name        = local.prefix
  description = "Security Group used by DonkeyVPN"
  vpc_id      = var.vpc_id

  tags = merge({
    Name = local.prefix
  }, local.base_tags)
}

resource "aws_vpc_security_group_ingress_rule" "allow_wireguard" {
  security_group_id = aws_security_group.default.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "udp"
  from_port         = 51820
  to_port           = 51820
}

resource "aws_vpc_security_group_ingress_rule" "allow_ssh" {
  security_group_id = aws_security_group.default.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "tcp"
  from_port         = 22
  to_port           = 22
}

resource "aws_vpc_security_group_egress_rule" "ipv4_traffic" {
  security_group_id = aws_security_group.default.id
  cidr_ipv4         = "0.0.0.0/0"
  ip_protocol       = "-1"
}

resource "aws_vpc_security_group_egress_rule" "ipv6_traffic" {
  security_group_id = aws_security_group.default.id
  cidr_ipv6         = "::/0"
  ip_protocol       = "-1"
}

resource "aws_launch_template" "default" {
  name = local.prefix

  iam_instance_profile {
    name = resource.aws_iam_instance_profile.this.name
  }

  image_id = data.aws_ami.debian.id

  instance_market_options {
    market_type = "spot"
  }

  instance_type = var.instance_type
  key_name      = var.key_name

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required"
    http_put_response_hop_limit = 1
    instance_metadata_tags      = "enabled"
  }

  monitoring {
    enabled = true
  }

  user_data = base64encode(templatefile("${path.module}/templates/userdata.tpl.sh", {
    in_vpn_record_ttl  = "60",
    in_domain_name     = var.vpn_domain_name
    in_hosted_zone_id  = var.hosted_zone
    in_ssm_private_key = var.private_key_ssm_param
    in_ssm_public_key  = var.public_key_ssm_param

    in_api_base_url    = "https://glorious-supposedly-lark.ngrok-free.app" # TODO: set http api endpoint
    in_api_secret      = var.api_secret
    in_use_route53     = var.hosted_zone != "" ? "true" : "false"
    # in_ssm_peers       = var.peers_ssm_param
  }))

  vpc_security_group_ids = [aws_security_group.default.id]
}

resource "aws_autoscaling_group" "default" {
  name                      = "${local.prefix}"
  max_size                  = var.max_size
  min_size                  = var.min_size
  desired_capacity          = var.desired_capacity
  health_check_grace_period = 300
  health_check_type         = "ELB"
  force_delete              = true
  termination_policies      = ["OldestInstance"]

  launch_template {
    id      = aws_launch_template.default.id
    version = "$Latest"
  }

  vpc_zone_identifier       = var.subnets

  instance_maintenance_policy {
    min_healthy_percentage = 90
    max_healthy_percentage = 120
  }

  dynamic "tag" {
    for_each = local.asg_tags
    content {
      key                 = tag.key
      value               = tag.value
      propagate_at_launch = true
    }
  }
}
