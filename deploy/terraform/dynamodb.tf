resource "aws_dynamodb_table" "donkeyvpn_peers" {
  name           = local.dynamodb_peers_table_name
  billing_mode   = "PAY_PER_REQUEST"

  hash_key       = "IPAddress"

  attribute {
    name = "IPAddress"
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

  hash_key       = "Id"

  attribute {
    name = "Id"
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
