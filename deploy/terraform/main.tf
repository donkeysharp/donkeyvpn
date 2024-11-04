locals {
  debian_ami_owner = "136693071363"

  default_prefix = "donkeyvpn-${var.environment}"
  prefix         = var.prefix != "" ? var.prefix : local.default_prefix

  ec2_role_name = "${local.prefix}-ec2"

  base_tags = {
    Project     = "donkeyvpn"
    Environment = var.environment
    ManagedBy   = "terraform"
  }

  asg_tags = merge({
    Name = local.prefix
  }, local.base_tags)
}

terraform {
  backend "s3" {
  }
}

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

resource "aws_iam_role" "this" {
  name = local.ec2_role_name

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })

  tags = merge(local.base_tags, {
    Name = local.ec2_role_name
  })
}

resource "aws_iam_instance_profile" "this" {
  name = aws_iam_role.this.name
  role = aws_iam_role.this.name
}

resource "aws_iam_policy" "vpn_permissions" {
  name        = "${local.ec2_role_name}-permissions"
  path        = "/"
  description = "Policy used by Donkey VPN project"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "route53:ChangeResourceRecordSets",
        ]
        Effect   = "Allow"
        Resource = "arn:aws:route53:::hostedzone/${var.hosted_zone}"
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "this" {
  role       = aws_iam_role.this.name
  policy_arn = aws_iam_policy.vpn_permissions.arn
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
    in_vpn_record_name = var.vpn_domain_name
    in_hosted_zone_id  = var.hosted_zone
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
