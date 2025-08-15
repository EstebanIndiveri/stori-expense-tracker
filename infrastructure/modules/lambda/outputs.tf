output "api_function_name" {
  description = "Name of the API Lambda function"
  value       = aws_lambda_function.api.function_name
}

output "api_function_arn" {
  description = "ARN of the API Lambda function"
  value       = aws_lambda_function.api.arn
}

output "api_function_invoke_arn" {
  description = "Invoke ARN of the API Lambda function"
  value       = aws_lambda_function.api.invoke_arn
}

output "ai_advisor_function_name" {
  description = "Name of the AI advisor Lambda function"
  value       = aws_lambda_function.ai_advisor.function_name
}

output "ai_advisor_function_arn" {
  description = "ARN of the AI advisor Lambda function"
  value       = aws_lambda_function.ai_advisor.arn
}

output "ai_advisor_function_invoke_arn" {
  description = "Invoke ARN of the AI advisor Lambda function"
  value       = aws_lambda_function.ai_advisor.invoke_arn
}

output "lambda_log_group_names" {
  description = "Names of the Lambda log groups"
  value = [
    aws_cloudwatch_log_group.api_lambda.name,
    aws_cloudwatch_log_group.ai_advisor_lambda.name
  ]
}

output "dead_letter_queue_arn" {
  description = "ARN of the dead letter queue"
  value       = aws_sqs_queue.lambda_dlq.arn
}
