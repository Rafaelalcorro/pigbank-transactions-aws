resource "aws_sqs_queue" "start_payment_dlq" {
  name                      = "start-payment-dlq"
  message_retention_seconds = 1209600
}

resource "aws_sqs_queue" "start_payment_queue" {
  name                       = "start-payment-queue"
  visibility_timeout_seconds = 30
  message_retention_seconds  = 86400

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.start_payment_dlq.arn
    maxReceiveCount     = 3
  })
}
