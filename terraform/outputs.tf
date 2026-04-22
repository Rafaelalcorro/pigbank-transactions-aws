output "api_url" {
  value = "${aws_api_gateway_stage.dev.invoke_url}/payment"
}

output "sqs_url" {
  value = aws_sqs_queue.start_payment_queue.url
}

output "sqs_arn" {
  value = aws_sqs_queue.start_payment_queue.arn
}
