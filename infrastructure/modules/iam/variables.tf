variable "app_name" {
  description = "Application name"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod."
  }
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "dynamodb_table_arn" {
  description = "ARN of the DynamoDB table"
  type        = string
}

variable "kms_key_arn" {
  description = "ARN of the KMS key for encryption"
  type        = string
}

variable "enable_vpc_access" {
  description = "Whether to enable VPC access for Lambda functions"
  type        = bool
  default     = false
}

variable "enable_frontend_deployment" {
  description = "Whether to enable frontend deployment resources"
  type        = bool
  default     = false
}

variable "frontend_s3_bucket_arn" {
  description = "ARN of the S3 bucket for frontend deployment"
  type        = string
  default     = ""
}

variable "cloudfront_distribution_arn" {
  description = "ARN of the CloudFront distribution"
  type        = string
  default     = ""
}

variable "enable_github_actions_role" {
  description = "Whether to create GitHub Actions deployment role"
  type        = bool
  default     = false
}

variable "github_actions_oidc_provider_arn" {
  description = "ARN of the GitHub Actions OIDC provider"
  type        = string
  default     = ""
}

variable "github_repository" {
  description = "GitHub repository in format 'owner/repo'"
  type        = string
  default     = ""
}

variable "tags" {
  description = "Common tags to apply to all resources"
  type        = map(string)
  default     = {}
}
