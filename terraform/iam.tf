data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "payment_lambda_role" {
  name               = "payment-lambda-role"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

resource "aws_iam_role_policy" "payment_lambda_policy" {
  name = "payment-lambda-policy"
  role = aws_iam_role.payment_lambda_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:*:*:*"
      },
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:Query"
        ]
        Resource = [
          "arn:aws:dynamodb:us-east-1:${var.account_id}:table/${var.card_table}",
          "arn:aws:dynamodb:us-east-1:${var.account_id}:table/${var.card_table}/index/*",
          "arn:aws:dynamodb:us-east-1:${var.account_id}:table/${var.payment_table}",
          "arn:aws:dynamodb:us-east-1:${var.account_id}:table/${var.payment_table}/index/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "sqs:SendMessage",
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes"
        ]
        Resource = [
          aws_sqs_queue.start_payment_queue.arn,
          aws_sqs_queue.start_payment_dlq.arn
        ]
      }
    ]
  })
}
