resource "aws_dynamodb_table" "donkeyvpn_peers" {
  name           = local.dynamodb_peers_table_name
  billing_mode   = "PAY_PER_REQUEST"

  hash_key       = "PeerAddress"
  range_key      = "PublicKey"

  attribute {
    name = "PeerAddress"
    type = "S"
  }

  attribute {
    name = "PublicKey"
    type = "S"
  }

  ttl {
    attribute_name = "TimeToExist"
    enabled        = true
  }

  tags = merge({
    Name = local.dynamodb_peers_table_name
  }, local.base_tags)
}

resource "aws_dynamodb_table" "donkeyvpn_instances" {
  name           = local.dynamodb_instances_table_name
  billing_mode   = "PAY_PER_REQUEST"

  hash_key       = "Hostname"
  range_key      = "PublicIP"

  attribute {
    name = "Hostname"
    type = "S"
  }

  attribute {
    name = "PublicIP"
    type = "S"
  }

  ttl {
    attribute_name = "TimeToExist"
    enabled        = true
  }

  tags = merge({
    Name = local.dynamodb_peers_table_name
  }, local.base_tags)
}
