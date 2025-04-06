
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
    name = resource.aws_iam_instance_profile.asg.name
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
    in_wg_interface_address = var.wireguard_interface_address
    in_wg_ip_range          = var.wireguard_ip_range

    in_vpn_record_ttl  = "60",
    in_domain_name     = var.vpn_domain_name
    in_hosted_zone_id  = var.hosted_zone
    in_ssm_private_key = var.private_key_ssm_param
    in_ssm_public_key  = var.public_key_ssm_param

    in_api_base_url    = trim(aws_apigatewayv2_stage.default.invoke_url, "/")
    in_api_secret      = data.aws_ssm_parameter.webhook_secret.value
    in_use_route53     = var.hosted_zone != "none" ? "true" : "false"
  }))

  vpc_security_group_ids = [aws_security_group.default.id]
}

resource "aws_autoscaling_group" "default" {
  name                      = local.autoscaling_group_name
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
