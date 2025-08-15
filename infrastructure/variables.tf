variable "aws_region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "stori-expense-tracker"
}

variable "domain_name" {
  description = "Domain name for the frontend"
  type        = string
  default     = ""
}

variable "api_domain_name" {
  description = "Domain name for the API"
  type        = string
  default     = ""
}

variable "certificate_arn" {
  description = "ACM certificate ARN for HTTPS"
  type        = string
  default     = ""
}

variable "cors_origins" {
  description = "Allowed CORS origins"
  type        = list(string)
  default     = ["*"]
}

variable "openai_api_key" {
  description = "OpenAI API key for AI advisor"
  type        = string
  sensitive   = true
}
