output "lambda_api_execution_role_arn" {
  description = "ARN of the Lambda API execution role"
  value       = aws_iam_role.lambda_api_execution.arn
}

output "lambda_ai_execution_role_arn" {
  description = "ARN of the Lambda AI execution role"
  value       = aws_iam_role.lambda_ai_execution.arn
}

output "lambda_api_execution_role_name" {
  description = "Name of the Lambda API execution role"
  value       = aws_iam_role.lambda_api_execution.name
}

output "lambda_ai_execution_role_name" {
  description = "Name of the Lambda AI execution role"
  value       = aws_iam_role.lambda_ai_execution.name
}

output "api_gateway_logging_role_arn" {
  description = "ARN of the API Gateway logging role"
  value       = aws_iam_role.api_gateway_logging.arn
}

output "lambda_api_dynamodb_policy_arn" {
  description = "ARN of the Lambda API DynamoDB policy"
  value       = aws_iam_policy.lambda_api_dynamodb.arn
}

output "lambda_ssm_parameters_policy_arn" {
  description = "ARN of the Lambda SSM parameters policy"
  value       = aws_iam_policy.lambda_ssm_parameters.arn
}

output "lambda_cloudwatch_logs_policy_arn" {
  description = "ARN of the Lambda CloudWatch Logs policy"
  value       = aws_iam_policy.lambda_cloudwatch_logs.arn
}

output "lambda_ai_dynamodb_read_policy_arn" {
  description = "ARN of the Lambda AI DynamoDB read-only policy"
  value       = aws_iam_policy.lambda_ai_dynamodb_read.arn
}

output "cloudfront_s3_access_policy_arn" {
  description = "ARN of the CloudFront S3 access policy"
  value       = var.enable_frontend_deployment ? aws_iam_policy.cloudfront_s3_access[0].arn : ""
}

output "github_actions_deployment_role_arn" {
  description = "ARN of the GitHub Actions deployment role"
  value       = var.enable_github_actions_role ? aws_iam_role.github_actions_deployment[0].arn : ""
}

output "github_actions_deployment_policy_arn" {
  description = "ARN of the GitHub Actions deployment policy"
  value       = var.enable_github_actions_role ? aws_iam_policy.github_actions_deployment[0].arn : ""
}

output "role_arns" {
  description = "Map of all IAM role ARNs"
  value = {
    lambda_api_execution      = aws_iam_role.lambda_api_execution.arn
    lambda_ai_execution       = aws_iam_role.lambda_ai_execution.arn
    api_gateway_logging       = aws_iam_role.api_gateway_logging.arn
    github_actions_deployment = var.enable_github_actions_role ? aws_iam_role.github_actions_deployment[0].arn : ""
  }
}

output "policy_arns" {
  description = "Map of all IAM policy ARNs"
  value = {
    lambda_api_dynamodb       = aws_iam_policy.lambda_api_dynamodb.arn
    lambda_ssm_parameters     = aws_iam_policy.lambda_ssm_parameters.arn
    lambda_cloudwatch_logs    = aws_iam_policy.lambda_cloudwatch_logs.arn
    lambda_ai_dynamodb_read   = aws_iam_policy.lambda_ai_dynamodb_read.arn
    cloudfront_s3_access      = var.enable_frontend_deployment ? aws_iam_policy.cloudfront_s3_access[0].arn : ""
    github_actions_deployment = var.enable_github_actions_role ? aws_iam_policy.github_actions_deployment[0].arn : ""
  }
}
