resource "aws_iam_role" "asg" {
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

resource "aws_iam_instance_profile" "asg" {
  name = aws_iam_role.asg.name
  role = aws_iam_role.asg.name
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
    ]
  })
}

resource "aws_iam_role_policy_attachment" "asg" {
  role       = aws_iam_role.asg.name
  policy_arn = aws_iam_policy.vpn_permissions.arn
}

# IAM resources required by DonkeyVPN running on lambda
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

resource "aws_iam_policy" "donkeyvpn_permissions" {
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
          "autoscaling:UpdateAutoScalingGroup",
          "autoscaling:TerminateInstanceInAutoScalingGroup",
        ]
        Effect   = "Allow"
        Resource = [
          aws_autoscaling_group.default.arn
        ]
      },
      {
        Action = [
          "autoscaling:DescribeAutoScalingGroups",
        ],
        Effect   = "Allow"
        Resource = "*"
      },
      {
        Action   = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ],
        Effect   = "Allow"
        Resource = "arn:aws:logs:*:*:*"
      },
      {
        Action   = [
          "ec2:DescribeInstances",
        ],
        Effect   = "Allow"
        Resource = "*"
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
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = aws_iam_policy.donkeyvpn_permissions.arn
}
