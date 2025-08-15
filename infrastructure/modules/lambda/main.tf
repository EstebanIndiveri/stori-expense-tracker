# Lambda Functions for Stori Backend
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# API Lambda Function
resource "aws_lambda_function" "api" {
  filename         = var.lambda_zip_path != "" ? var.lambda_zip_path : "${path.module}/../../packages/backend/bin/api.zip"
  function_name    = "${var.project_name}-api-${var.environment}"
  role            = var.lambda_execution_role_arn
  handler         = "api"
  runtime         = "go1.x"
  timeout         = 30
  memory_size     = 512
  
  # Lambda layers for better performance
  layers = var.lambda_layers
  
  environment {
    variables = {
      ENVIRONMENT           = var.environment
      DYNAMODB_TABLE_NAME   = var.dynamodb_table_name
      OPENAI_API_KEY_SSM    = var.openai_api_key_ssm
      LOG_LEVEL            = var.log_level
      CORS_ORIGINS         = join(",", var.cors_origins)
    }
  }
  
  # VPC configuration for enhanced security
  dynamic "vpc_config" {
    for_each = var.vpc_config != null ? [var.vpc_config] : []
    content {
      subnet_ids         = vpc_config.value.subnet_ids
      security_group_ids = vpc_config.value.security_group_ids
    }
  }
  
  # Enable X-Ray tracing
  tracing_config {
    mode = var.enable_xray ? "Active" : "PassThrough"
  }
  
  # Dead letter queue for failed invocations
  dead_letter_config {
    target_arn = aws_sqs_queue.lambda_dlq.arn
  }
  
  tags = merge(var.tags, {
    Name = "${var.project_name}-api-${var.environment}"
    Type = "api"
  })
  
  depends_on = [
    aws_iam_role_policy_attachment.lambda_logs,
    aws_cloudwatch_log_group.api_lambda,
  ]
}

# AI Advisor Lambda Function
resource "aws_lambda_function" "ai_advisor" {
  filename         = var.lambda_zip_path != "" ? var.lambda_ai_zip_path : "${path.module}/../../packages/backend/bin/ai-advisor.zip"
  function_name    = "${var.project_name}-ai-advisor-${var.environment}"
  role            = var.lambda_execution_role_arn
  handler         = "ai-advisor"
  runtime         = "go1.x"
  timeout         = 60  # AI calls may take longer
  memory_size     = 1024 # More memory for AI processing
  
  environment {
    variables = {
      ENVIRONMENT           = var.environment
      DYNAMODB_TABLE_NAME   = var.dynamodb_table_name
      OPENAI_API_KEY_SSM    = var.openai_api_key_ssm
      LOG_LEVEL            = var.log_level
    }
  }
  
  # VPC configuration
  dynamic "vpc_config" {
    for_each = var.vpc_config != null ? [var.vpc_config] : []
    content {
      subnet_ids         = vpc_config.value.subnet_ids
      security_group_ids = vpc_config.value.security_group_ids
    }
  }
  
  # Enable X-Ray tracing
  tracing_config {
    mode = var.enable_xray ? "Active" : "PassThrough"
  }
  
  # Dead letter queue
  dead_letter_config {
    target_arn = aws_sqs_queue.lambda_dlq.arn
  }
  
  tags = merge(var.tags, {
    Name = "${var.project_name}-ai-advisor-${var.environment}"
    Type = "ai-advisor"
  })
  
  depends_on = [
    aws_iam_role_policy_attachment.lambda_logs,
    aws_cloudwatch_log_group.ai_advisor_lambda,
  ]
}

# CloudWatch Log Groups
resource "aws_cloudwatch_log_group" "api_lambda" {
  name              = "/aws/lambda/${var.project_name}-api-${var.environment}"
  retention_in_days = var.log_retention_days
  kms_key_id        = var.kms_key_arn
  
  tags = var.tags
}

resource "aws_cloudwatch_log_group" "ai_advisor_lambda" {
  name              = "/aws/lambda/${var.project_name}-ai-advisor-${var.environment}"
  retention_in_days = var.log_retention_days
  kms_key_id        = var.kms_key_arn
  
  tags = var.tags
}

# Dead Letter Queue for failed Lambda invocations
resource "aws_sqs_queue" "lambda_dlq" {
  name                       = "${var.project_name}-lambda-dlq-${var.environment}"
  message_retention_seconds  = 1209600  # 14 days
  visibility_timeout_seconds = 60
  
  # Enable encryption
  kms_master_key_id = var.kms_key_arn
  
  tags = merge(var.tags, {
    Name = "${var.project_name}-lambda-dlq-${var.environment}"
    Purpose = "Dead letter queue for failed Lambda invocations"
  })
}

# Lambda permission for API Gateway to invoke API function
resource "aws_lambda_permission" "api_gateway_invoke_api" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.api.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${var.api_gateway_execution_arn}/*/*"
}

# Lambda permission for API Gateway to invoke AI advisor function
resource "aws_lambda_permission" "api_gateway_invoke_ai" {
  statement_id  = "AllowAPIGatewayInvokeAI"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.ai_advisor.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${var.api_gateway_execution_arn}/*/*"
}

# CloudWatch Alarms for monitoring
resource "aws_cloudwatch_metric_alarm" "api_lambda_errors" {
  alarm_name          = "${var.project_name}-api-lambda-errors-${var.environment}"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = "300"
  statistic           = "Sum"
  threshold           = "5"
  alarm_description   = "This metric monitors API Lambda errors"
  alarm_actions       = var.sns_alarm_topic_arn != "" ? [var.sns_alarm_topic_arn] : []
  
  dimensions = {
    FunctionName = aws_lambda_function.api.function_name
  }
  
  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "api_lambda_duration" {
  alarm_name          = "${var.project_name}-api-lambda-duration-${var.environment}"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "Duration"
  namespace           = "AWS/Lambda"
  period              = "300"
  statistic           = "Average"
  threshold           = "25000"  # 25 seconds
  alarm_description   = "This metric monitors API Lambda duration"
  alarm_actions       = var.sns_alarm_topic_arn != "" ? [var.sns_alarm_topic_arn] : []
  
  dimensions = {
    FunctionName = aws_lambda_function.api.function_name
  }
  
  tags = var.tags
}

# Lambda provisioned concurrency for production
resource "aws_lambda_provisioned_concurrency_config" "api_lambda_concurrency" {
  count                     = var.environment == "prod" ? 1 : 0
  function_name             = aws_lambda_function.api.function_name
  provisioned_concurrency_units = var.provisioned_concurrency_units
  qualifier                 = aws_lambda_function.api.version
}
