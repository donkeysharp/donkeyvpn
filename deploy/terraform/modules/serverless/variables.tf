variable "lambda_role_name" {
  type        = string
  description = "IAM role name that will be assigned to lambda function"
}

variable "golang_executable_path" {
  type        = string
  description = "Where the 'bootstrap' executable is located"
}

variable "lambda_function_name" {
  type        = string
  description = "Lambda function name for donkeyvpn"
}

variable "apigateway_name" {
  type        = string
  description = "API Gateway name"
}

variable "webhook_secret" {
  type        = string
  description = "Webhook secret value"
}

variable "telegram_bot_api_token" {
  type        = string
  description = "Telegram API token"
}

variable "autoscaling_group_name" {
  type        = string
  description = "Autoscaling Group name. This ASG is the one that will contain the VPN instances"
}

variable "dynamodb_peers_table_name" {
  type        = string
  description = "Dynamodb table where peer information is going to be stored"
}

variable "dynamodb_instances_table_name" {
  type        = string
  description = "Dynamodb table where instance information is going to be stored"
}
