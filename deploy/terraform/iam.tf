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
      {
        Action = [
          "ssm:GetParameter"
        ]
        Effect   = "Allow"
        Resource = [
          "arn:aws:ssm:${local.region}:${local.account_id}:parameter/*"
        ]
      },
      {
        Action = [
          "kms:Decrypt"
        ]
        Effect   = "Allow"
        Resource = [
          data.aws_kms_alias.ssm.target_key_arn
        ]
      },
      {
        Action = [
          "dynamodb:PutItem",
          "dynamodb:GetItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Query",
          "dynamodb:Scan"
        ]
        Effect   = "Allow"
        Resource = [
          aws_dynamodb_table.donkeyvpn_peers.arn,
          aws_dynamodb_table.donkeyvpn_instances.arn
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "this" {
  role       = aws_iam_role.this.name
  policy_arn = aws_iam_policy.vpn_permissions.arn
}

resource "aws_iam_role" "lambda_exec" {
  name = "${local.prefix}-lambda-exec"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action    = "sts:AssumeRole"
        Effect    = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_basic_execution" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_policy" "lambda_permissions" {
  name        = "${local.prefix}-lambda-exec-permissions"
  path        = "/"
  description = "Policy used by Donkey VPN project in lambda"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "dynamodb:PutItem",
          "dynamodb:GetItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Query",
          "dynamodb:Scan"
        ]
        Effect   = "Allow"
        Resource = [
          aws_dynamodb_table.donkeyvpn_peers.arn,
          aws_dynamodb_table.donkeyvpn_instances.arn
        ]
      },
      {
        Action = [
          "ec2:*"
        ]
        Effect   = "Allow"
        Resource = [
          "*"
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = aws_iam_policy.lambda_permissions.arn
}
