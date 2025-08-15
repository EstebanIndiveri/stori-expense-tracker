output "openai_api_key_parameter_name" {
  description = "Name of the OpenAI API key parameter"
  value       = aws_ssm_parameter.openai_api_key.name
}

output "openai_api_key_parameter_arn" {
  description = "ARN of the OpenAI API key parameter"
  value       = aws_ssm_parameter.openai_api_key.arn
}

output "dynamodb_table_name_parameter_name" {
  description = "Name of the DynamoDB table name parameter"
  value       = aws_ssm_parameter.dynamodb_table_name.name
}

output "api_gateway_url_parameter_name" {
  description = "Name of the API Gateway URL parameter"
  value       = aws_ssm_parameter.api_gateway_url.name
}

output "frontend_url_parameter_name" {
  description = "Name of the frontend URL parameter"
  value       = var.frontend_url != "" ? aws_ssm_parameter.frontend_url[0].name : ""
}

output "log_level_parameter_name" {
  description = "Name of the log level parameter"
  value       = aws_ssm_parameter.log_level.name
}

output "aws_region_parameter_name" {
  description = "Name of the AWS region parameter"
  value       = aws_ssm_parameter.aws_region.name
}

output "ai_prompt_template_parameter_name" {
  description = "Name of the AI prompt template parameter"
  value       = aws_ssm_parameter.ai_prompt_template.name
}

output "rate_limit_parameter_name" {
  description = "Name of the rate limit parameter"
  value       = aws_ssm_parameter.rate_limit_per_minute.name
}

output "parameter_names" {
  description = "Map of all parameter names"
  value = {
    openai_api_key      = aws_ssm_parameter.openai_api_key.name
    dynamodb_table_name = aws_ssm_parameter.dynamodb_table_name.name
    api_gateway_url     = aws_ssm_parameter.api_gateway_url.name
    frontend_url        = var.frontend_url != "" ? aws_ssm_parameter.frontend_url[0].name : ""
    log_level           = aws_ssm_parameter.log_level.name
    aws_region          = aws_ssm_parameter.aws_region.name
    ai_prompt_template  = aws_ssm_parameter.ai_prompt_template.name
    rate_limit          = aws_ssm_parameter.rate_limit_per_minute.name
  }
}

output "parameter_arns" {
  description = "Map of all parameter ARNs for IAM policies"
  value = {
    openai_api_key      = aws_ssm_parameter.openai_api_key.arn
    dynamodb_table_name = aws_ssm_parameter.dynamodb_table_name.arn
    api_gateway_url     = aws_ssm_parameter.api_gateway_url.arn
    frontend_url        = var.frontend_url != "" ? aws_ssm_parameter.frontend_url[0].arn : ""
    log_level           = aws_ssm_parameter.log_level.arn
    aws_region          = aws_ssm_parameter.aws_region.arn
    ai_prompt_template  = aws_ssm_parameter.ai_prompt_template.arn
    rate_limit          = aws_ssm_parameter.rate_limit_per_minute.arn
  }
}

output "log_group_name" {
  description = "Name of the SSM access log group"
  value       = aws_cloudwatch_log_group.ssm_access.name
}

output "log_group_arn" {
  description = "ARN of the SSM access log group"
  value       = aws_cloudwatch_log_group.ssm_access.arn
}
