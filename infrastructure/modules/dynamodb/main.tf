# DynamoDB Table for Transactions
resource "aws_dynamodb_table" "transactions" {
  name           = "${var.project_name}-transactions-${var.environment}"
  billing_mode   = "ON_DEMAND"
  hash_key       = "pk"
  range_key      = "sk"
  
  # Enable encryption at rest
  server_side_encryption {
    enabled     = true
    kms_key_id  = aws_kms_key.dynamodb_key.arn
  }
  
  # Enable point-in-time recovery
  point_in_time_recovery {
    enabled = true
  }
  
  # Enable deletion protection for production
  deletion_protection_enabled = var.environment == "prod"
  
  # Primary key attributes
  attribute {
    name = "pk"
    type = "S"
  }
  
  attribute {
    name = "sk"
    type = "S"
  }
  
  # GSI1 attributes (Category + Month access)
  attribute {
    name = "gsi1pk"
    type = "S"
  }
  
  attribute {
    name = "gsi1sk"
    type = "S"
  }
  
  # GSI2 attributes (Type + Month access)
  attribute {
    name = "gsi2pk"
    type = "S"
  }
  
  attribute {
    name = "gsi2sk"
    type = "S"
  }
  
  # GSI1: Category + Month access pattern
  global_secondary_index {
    name            = "GSI1-CategoryMonth"
    hash_key        = "gsi1pk"
    range_key       = "gsi1sk"
    projection_type = "INCLUDE"
    non_key_attributes = [
      "amount", "category", "type", "date", "description", "txId"
    ]
  }
  
  # GSI2: Type + Month access pattern
  global_secondary_index {
    name            = "GSI2-TypeMonth"
    hash_key        = "gsi2pk"
    range_key       = "gsi2sk"
    projection_type = "INCLUDE"
    non_key_attributes = [
      "amount", "category", "type", "date", "description", "txId"
    ]
  }
  
  tags = {
    Name        = "${var.project_name}-transactions-${var.environment}"
    Environment = var.environment
    Purpose     = "Transaction storage with optimized access patterns"
  }
}

# KMS Key for DynamoDB encryption
resource "aws_kms_key" "dynamodb_key" {
  description             = "${var.project_name} DynamoDB encryption key - ${var.environment}"
  deletion_window_in_days = var.environment == "prod" ? 30 : 7
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "Enable IAM User Permissions"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:root"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "Allow DynamoDB Service"
        Effect = "Allow"
        Principal = {
          Service = "dynamodb.amazonaws.com"
        }
        Action = [
          "kms:Encrypt",
          "kms:Decrypt",
          "kms:ReEncrypt*",
          "kms:GenerateDataKey*",
          "kms:CreateGrant",
          "kms:DescribeKey"
        ]
        Resource = "*"
      }
    ]
  })
  
  tags = {
    Name        = "${var.project_name}-dynamodb-key-${var.environment}"
    Environment = var.environment
  }
}

resource "aws_kms_alias" "dynamodb_key_alias" {
  name          = "alias/${var.project_name}-dynamodb-${var.environment}"
  target_key_id = aws_kms_key.dynamodb_key.key_id
}

# Data source for current AWS account
data "aws_caller_identity" "current" {}
