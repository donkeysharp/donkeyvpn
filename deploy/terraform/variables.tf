variable "hosted_zone" {
  type        = string
  description = "Hosted Zone ID for the VPN"
  default     = "none"
}

variable "environment" {
  type        = string
  description = "Environment name (mainly used in tags)"
}

variable "prefix" {
  type        = string
  description = "Prefix to be used by all resources"
  default     = ""
}

variable "instance_type" {
  type = string
}

variable "key_name" {
  type = string
}

variable "vpn_domain_name" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "subnets" {
  type = list(string)
}

variable "max_size" {
  type = number
  default = 5
}

variable "min_size" {
  type = number
  default= 0
}

variable "desired_capacity" {
  type = number
  default = 0
}

variable "private_key_ssm_param" {
  type = string
  default = "/donkeyvpn/privatekey"
  description = "It is expected that SSM parameter is a SecureString that uses a KMS to encrypt/decrypt. It could be the default one"
}

variable "public_key_ssm_param" {
  type = string
  default = "/donkeyvpn/publickey"
  description = "It is expected that SSM parameter is a SecureString that uses a KMS to encrypt/decrypt. It could be the default one"
}

variable "telegram_bot_api_token_ssm_param" {
  type = string
  default = "/donkeyvpn/telegrambotapikey"
  description = "SSM Parameter where telegram bot api token is located"
}

variable "webhook_secret_ssm_param" {
  type = string
  default = "/donkeyvpn/webhooksecret"
  description = "SSM PArameter where webhook secret is located"
}

variable "kms_key_alias" {
  type = string
  default = "alias/aws/ssm"
}

variable "golang_executable_path" {
  type        = string
  description = "Where the 'bootstrap' executable is located"
  default     = "../../dist"
}

variable "wireguard_interface_address" {
  type        = string
  description = "(optional) IP address and range for wireguard VPN server"
  default     = "10.0.0.1/24"
}

variable "wireguard_ip_range" {
  type        = string
  description = "(optional) IP address range for wireguard VPN server"
  default     = "10.0.0.0/24"
}

variable "testing_userdata_api_base_url" {
  type        = string
  description = "(optional) Use this only when running in development mode, this will make VPN instances to call this url to register itself once ready."
  default     = ""
}

variable "notifier_rate" {
  type        = string
  description = "(optional) Rate in which notifier will be executed"
  default     = "cron(*/30 * * * ? *)"
}
