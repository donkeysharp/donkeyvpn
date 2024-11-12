variable "hosted_zone" {
  type        = string
  description = "Hosted Zone ID for the VPN"
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
  default = 0
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
  default = "donkeyvpn/privatekey"
  description = "It is expected that SSM parameter is a SecureString that uses a KMS to encrypt/decrypt. It could be the default one"
}

variable "public_key_ssm_param" {
  type = string
  default = "donkeyvpn/publickey"
  description = "It is expected that SSM parameter is a SecureString that uses a KMS to encrypt/decrypt. It could be the default one"
}

variable "kms_key_alias" {
  type = string
  default = "alias/aws/ssm"
}
