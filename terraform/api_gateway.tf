resource "aws_api_gateway_rest_api" "payment_api" {
  name        = "payment-service-api"
  description = "API para el microservicio de pagos"
}

resource "aws_api_gateway_resource" "payment" {
  rest_api_id = aws_api_gateway_rest_api.payment_api.id
  parent_id   = aws_api_gateway_rest_api.payment_api.root_resource_id
  path_part   = "payment"
}

resource "aws_api_gateway_method" "post_payment" {
  rest_api_id   = aws_api_gateway_rest_api.payment_api.id
  resource_id   = aws_api_gateway_resource.payment.id
  http_method   = "POST"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "post_payment" {
  rest_api_id             = aws_api_gateway_rest_api.payment_api.id
  resource_id             = aws_api_gateway_resource.payment.id
  http_method             = aws_api_gateway_method.post_payment.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.post_payment.invoke_arn

  depends_on = [aws_api_gateway_method.post_payment]
}


resource "aws_lambda_permission" "apigw_post_payment" {
  statement_id  = "AllowAPIGatewayInvokePayment"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.post_payment.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.payment_api.execution_arn}/*/*"
}

# Deployment
resource "aws_api_gateway_deployment" "payment_deploy" {
  rest_api_id = aws_api_gateway_rest_api.payment_api.id

  depends_on = [
    aws_api_gateway_integration.post_payment,
  ]

  lifecycle {
    create_before_destroy = true
  }
}

# Stage dev
resource "aws_api_gateway_stage" "dev" {
  deployment_id = aws_api_gateway_deployment.payment_deploy.id
  rest_api_id   = aws_api_gateway_rest_api.payment_api.id
  stage_name    = "dev"
}

# Recurso
resource "aws_api_gateway_resource" "payment_trace" {
  rest_api_id = aws_api_gateway_rest_api.payment_api.id
  parent_id   = aws_api_gateway_resource.payment.id
  path_part   = "{traceId}"
}

# GET
resource "aws_api_gateway_method" "get_payment_status" {
  rest_api_id   = aws_api_gateway_rest_api.payment_api.id
  resource_id   = aws_api_gateway_resource.payment_trace.id
  http_method   = "GET"
  authorization = "NONE"
}

# Integraciion
resource "aws_api_gateway_integration" "get_payment_status" {
  rest_api_id             = aws_api_gateway_rest_api.payment_api.id
  resource_id             = aws_api_gateway_resource.payment_trace.id
  http_method             = aws_api_gateway_method.get_payment_status.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.get_payment_status.invoke_arn

  depends_on = [aws_api_gateway_method.get_payment_status]
}

# Permiso
resource "aws_lambda_permission" "apigw_get_payment_status" {
  statement_id  = "AllowAPIGatewayInvokeGetStatus"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.get_payment_status.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.payment_api.execution_arn}/*/*"
}

# Redeploy
resource "aws_api_gateway_deployment" "payment_deploy_v2" {
  rest_api_id = aws_api_gateway_rest_api.payment_api.id

  depends_on = [
    aws_api_gateway_integration.post_payment,
    aws_api_gateway_integration.get_payment_status,
  ]

  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_resource.payment.id,
      aws_api_gateway_resource.payment_trace.id,
      aws_api_gateway_method.get_payment_status.id,
      aws_api_gateway_integration.get_payment_status.id,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "dev_v2" {
  deployment_id = aws_api_gateway_deployment.payment_deploy_v2.id
  rest_api_id   = aws_api_gateway_rest_api.payment_api.id
  stage_name    = "dev"

  depends_on = [aws_api_gateway_stage.dev]
}


# Recurso /payment/user/{userId}
resource "aws_api_gateway_resource" "payment_user" {
  rest_api_id = aws_api_gateway_rest_api.payment_api.id
  parent_id   = aws_api_gateway_resource.payment.id
  path_part   = "user"
}

resource "aws_api_gateway_resource" "payment_user_id" {
  rest_api_id = aws_api_gateway_rest_api.payment_api.id
  parent_id   = aws_api_gateway_resource.payment_user.id
  path_part   = "{userId}"
}

# Método GET /payment/user/{userId}
resource "aws_api_gateway_method" "get_user_payments" {
  rest_api_id   = aws_api_gateway_rest_api.payment_api.id
  resource_id   = aws_api_gateway_resource.payment_user_id.id
  http_method   = "GET"
  authorization = "NONE"
}

# Integración
resource "aws_api_gateway_integration" "get_user_payments" {
  rest_api_id             = aws_api_gateway_rest_api.payment_api.id
  resource_id             = aws_api_gateway_resource.payment_user_id.id
  http_method             = aws_api_gateway_method.get_user_payments.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.get_user_payments.invoke_arn

  depends_on = [aws_api_gateway_method.get_user_payments]
}

# Permiso
resource "aws_lambda_permission" "apigw_get_user_payments" {
  statement_id  = "AllowAPIGatewayInvokeGetUserPayments"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.get_user_payments.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.payment_api.execution_arn}/*/*"
}

# Redeploy v3
resource "aws_api_gateway_deployment" "payment_deploy_v3" {
  rest_api_id = aws_api_gateway_rest_api.payment_api.id

  depends_on = [
    aws_api_gateway_integration.post_payment,
    aws_api_gateway_integration.get_payment_status,
    aws_api_gateway_integration.get_user_payments,
  ]

  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_resource.payment_user_id.id,
      aws_api_gateway_method.get_user_payments.id,
      aws_api_gateway_integration.get_user_payments.id,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "dev_v3" {
  deployment_id = aws_api_gateway_deployment.payment_deploy_v3.id
  rest_api_id   = aws_api_gateway_rest_api.payment_api.id
  stage_name    = "dev"

  depends_on = [aws_api_gateway_stage.dev_v2]
}
