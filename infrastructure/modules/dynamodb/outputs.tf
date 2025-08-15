output "table_name" {
  description = "Name of the DynamoDB table"
  value       = aws_dynamodb_table.transactions.name
}

output "table_arn" {
  description = "ARN of the DynamoDB table"
  value       = aws_dynamodb_table.transactions.arn
}

output "table_id" {
  description = "ID of the DynamoDB table"
  value       = aws_dynamodb_table.transactions.id
}

output "kms_key_arn" {
  description = "ARN of the KMS key"
  value       = aws_kms_key.dynamodb_key.arn
}

output "gsi_names" {
  description = "Names of Global Secondary Indexes"
  value = [
    "GSI1-CategoryMonth",
    "GSI2-TypeMonth"
  ]
}
