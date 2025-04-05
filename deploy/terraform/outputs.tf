output "apigateway_url" {
  value = trim(aws_apigatewayv2_stage.default.invoke_url, "/")
}
