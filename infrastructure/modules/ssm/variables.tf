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

variable "openai_api_key" {
  description = "OpenAI API key for AI advisor functionality"
  type        = string
  sensitive   = true
}

variable "dynamodb_table_name" {
  description = "DynamoDB table name for expense tracker"
  type        = string
}

variable "api_gateway_url" {
  description = "API Gateway URL for the expense tracker API"
  type        = string
}

variable "frontend_url" {
  description = "Frontend URL for CORS configuration"
  type        = string
  default     = ""
}

variable "log_level" {
  description = "Logging level for Lambda functions"
  type        = string
  default     = "INFO"
  validation {
    condition     = contains(["DEBUG", "INFO", "WARN", "ERROR"], var.log_level)
    error_message = "Log level must be DEBUG, INFO, WARN, or ERROR."
  }
}

variable "aws_region" {
  description = "AWS region for resource configuration"
  type        = string
}

variable "ai_prompt_template" {
  description = "AI prompt template for financial advice"
  type        = string
  default     = <<EOF
You are a financial advisor AI. Based on the user's expense data provided below, give practical and personalized financial advice.

Context:
- User's monthly expenses and income data
- Categories with highest spending
- Recent transaction patterns

Please provide:
1. Spending analysis summary
2. Areas for potential savings
3. Budget recommendations
4. Specific actionable advice

User Data: {data}

Provide a helpful, concise response in a friendly tone.
EOF
}

variable "rate_limit_per_minute" {
  description = "API rate limit per minute"
  type        = number
  default     = 100
}

variable "kms_key_id" {
  description = "KMS key ID for parameter encryption"
  type        = string
}

variable "log_retention_days" {
  description = "CloudWatch log retention in days"
  type        = number
  default     = 30
}

variable "ssm_access_threshold" {
  description = "Threshold for SSM parameter access alarm"
  type        = number
  default     = 1000
}

variable "alarm_actions" {
  description = "List of ARNs to notify when alarm triggers"
  type        = list(string)
  default     = []
}

variable "tags" {
  description = "Common tags to apply to all resources"
  type        = map(string)
  default     = {}
}
