resource "aws_lambda_function" "post_payment" {
  function_name = "post-payment-lambda"
  role          = aws_iam_role.payment_lambda_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  filename      = "${path.module}/../lambdas/post-payment/post-payment.zip"
  timeout       = 15

  environment {
    variables = {
      CARD_TABLE        = var.card_table
      PAYMENT_TABLE     = var.payment_table
      START_PAYMENT_SQS = aws_sqs_queue.start_payment_queue.url
      REGION            = var.region
    }
  }

  depends_on = [aws_iam_role_policy.payment_lambda_policy]
}

resource "aws_lambda_function" "start_payment" {
  function_name = "start-payment-lambda"
  role          = aws_iam_role.payment_lambda_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  filename      = "${path.module}/../lambdas/start-payment/start-payment.zip"
  timeout       = 30

  environment {
    variables = {
      PAYMENT_TABLE = var.payment_table
      REGION        = var.region
    }
  }

  depends_on = [aws_iam_role_policy.payment_lambda_policy]
}

resource "aws_lambda_event_source_mapping" "sqs_to_start_payment" {
  event_source_arn = aws_sqs_queue.start_payment_queue.arn
  function_name    = aws_lambda_function.start_payment.arn
  batch_size       = 1
  enabled          = true
}


resource "aws_lambda_function" "get_payment_status" {
  function_name = "get-payment-status-lambda"
  role          = aws_iam_role.payment_lambda_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  filename      = "${path.module}/../lambdas/get-payment-status/get-payment-status.zip"
  timeout       = 15

  environment {
    variables = {
      PAYMENT_TABLE = var.payment_table
      REGION        = var.region
    }
  }

  depends_on = [aws_iam_role_policy.payment_lambda_policy]
}
