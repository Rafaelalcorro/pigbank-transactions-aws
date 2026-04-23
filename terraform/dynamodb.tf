resource "aws_dynamodb_table" "payment_gsi" {
  name         = "payment"
  billing_mode = "PAY_PER_REQUEST"

  hash_key = "traceId"

  attribute {
    name = "traceId"
    type = "S"
  }

  attribute {
    name = "userId"
    type = "S"
  }

  global_secondary_index {
    name            = "userId-index"
    hash_key        = "userId"
    projection_type = "ALL"
  }

  lifecycle {
    ignore_changes = [
      billing_mode,
      read_capacity,
      write_capacity,
    ]
  }
}
