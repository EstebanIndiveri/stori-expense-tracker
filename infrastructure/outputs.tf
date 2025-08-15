output "dynamodb_table_name" {
  description = "Name of the DynamoDB table"
  value       = module.dynamodb.table_name
}

output "dynamodb_table_arn" {
  description = "ARN of the DynamoDB table"
  value       = module.dynamodb.table_arn
}

output "api_gateway_url" {
  description = "URL of the API Gateway"
  value       = module.api_gateway.api_url
}

output "frontend_bucket_name" {
  description = "Name of the S3 bucket for frontend"
  value       = module.frontend.bucket_name
}

output "frontend_cloudfront_domain" {
  description = "CloudFront distribution domain for frontend"
  value       = module.frontend.cloudfront_domain
}

output "frontend_url" {
  description = "Frontend URL"
  value       = var.domain_name != "" ? "https://${var.domain_name}" : "https://${module.frontend.cloudfront_domain}"
}

output "lambda_function_names" {
  description = "Names of Lambda functions"
  value = {
    api        = module.lambda.api_function_name
    ai_advisor = module.lambda.ai_advisor_function_name
  }
}
