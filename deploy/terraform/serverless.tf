
data "archive_file" "binary" {
  type        = "zip"
  source_file = "${var.golang_executable_path}/bootstrap"
  output_path = "${var.golang_executable_path}/bootstrap.zip"
}

resource "aws_lambda_function" "golang_function" {
  function_name = "${local.default_prefix}-echo-api"
  role          = aws_iam_role.lambda_exec.arn
  handler       = "bootstrap" # it must be bootstrap

  filename      = "${var.golang_executable_path}/bootstrap.zip"
  source_code_hash = data.archive_file.binary.output_base64sha256

  runtime = "provided.al2023"

  environment {
    variables = {
      RUN_AS_LAMBDA                 = "true"
      WEBHOOK_SECRET                = var.api_secret
      TELEGRAM_BOT_API_TOKEN        = var.telegram_bot_api_token
      AUTOSCALING_GROUP_NAME        = local.autoscaling_group_name
      DYNAMODB_PEERS_TABLE_NAME     = aws_dynamodb_table.donkeyvpn_peers.name
      DYNAMODB_INSTANCES_TABLE_NAME = aws_dynamodb_table.donkeyvpn_instances.name
    }
  }
}

resource "aws_apigatewayv2_api" "default" {
  name          = "${local.default_prefix}-echo-api"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_route" "all_routes" {
  api_id    = aws_apigatewayv2_api.default.id
  route_key = "$default"
  target = "integrations/${aws_apigatewayv2_integration.lambda_integration.id}"
}

resource "aws_apigatewayv2_integration" "lambda_integration" {
  api_id           = aws_apigatewayv2_api.default.id
  integration_type = "AWS_PROXY"

  connection_type           = "INTERNET"
  description               = "Lambda integration used by donkeyvpn"
  integration_method        = "POST"
  integration_uri           = aws_lambda_function.golang_function.invoke_arn
  payload_format_version    = "2.0"
}

resource "aws_lambda_permission" "lambda_permission_http_api" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.golang_function.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.default.execution_arn}/*"
}

resource "aws_apigatewayv2_stage" "default" {
  api_id = aws_apigatewayv2_api.default.id
  name   = "$default"
  auto_deploy = true
}
