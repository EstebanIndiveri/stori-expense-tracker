resource "aws_ssm_parameter" "openai_api_key" {
  name        = "/${var.app_name}/${var.environment}/openai/api_key"
  type        = "SecureString"
  value       = var.openai_api_key
  description = "OpenAI API key for AI advisor functionality"
  key_id      = var.kms_key_id

  tags = merge(var.tags, {
    Name        = "${var.app_name}-${var.environment}-openai-api-key"
    Component   = "configuration"
    SecretType  = "api-key"
  })
}

resource "aws_ssm_parameter" "dynamodb_table_name" {
  name        = "/${var.app_name}/${var.environment}/dynamodb/table_name"
  type        = "String"
  value       = var.dynamodb_table_name
  description = "DynamoDB table name for expense tracker"

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-dynamodb-table-name"
    Component = "configuration"
  })
}

resource "aws_ssm_parameter" "api_gateway_url" {
  name        = "/${var.app_name}/${var.environment}/api/gateway_url"
  type        = "String"
  value       = var.api_gateway_url
  description = "API Gateway URL for the expense tracker API"

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-api-gateway-url"
    Component = "configuration"
  })
}

resource "aws_ssm_parameter" "frontend_url" {
  count       = var.frontend_url != "" ? 1 : 0
  name        = "/${var.app_name}/${var.environment}/frontend/url"
  type        = "String"
  value       = var.frontend_url
  description = "Frontend URL for CORS configuration"

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-frontend-url"
    Component = "configuration"
  })
}

resource "aws_ssm_parameter" "log_level" {
  name        = "/${var.app_name}/${var.environment}/logging/level"
  type        = "String"
  value       = var.log_level
  description = "Logging level for Lambda functions"

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-log-level"
    Component = "configuration"
  })
}

resource "aws_ssm_parameter" "aws_region" {
  name        = "/${var.app_name}/${var.environment}/aws/region"
  type        = "String"
  value       = var.aws_region
  description = "AWS region for resource configuration"

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-aws-region"
    Component = "configuration"
  })
}

# Parameter for AI prompt configuration
resource "aws_ssm_parameter" "ai_prompt_template" {
  name        = "/${var.app_name}/${var.environment}/ai/prompt_template"
  type        = "String"
  value       = var.ai_prompt_template
  description = "AI prompt template for financial advice"

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-ai-prompt-template"
    Component = "configuration"
  })
}

# Parameter for rate limiting configuration
resource "aws_ssm_parameter" "rate_limit_per_minute" {
  name        = "/${var.app_name}/${var.environment}/api/rate_limit_per_minute"
  type        = "String"
  value       = tostring(var.rate_limit_per_minute)
  description = "API rate limit per minute"

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-rate-limit"
    Component = "configuration"
  })
}

# CloudWatch Log Group for parameter access logging
resource "aws_cloudwatch_log_group" "ssm_access" {
  name              = "/aws/ssm/${var.app_name}/${var.environment}/access"
  retention_in_days = var.log_retention_days
  kms_key_id        = var.kms_key_id

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-ssm-access-logs"
    Component = "monitoring"
  })
}

# CloudWatch Metric Filter for parameter access monitoring
resource "aws_cloudwatch_metric_filter" "ssm_parameter_access" {
  name           = "${var.app_name}-${var.environment}-ssm-parameter-access"
  log_group_name = aws_cloudwatch_log_group.ssm_access.name
  pattern        = "[timestamp, request_id, event_type=\"Parameter\", ...]"

  metric_transformation {
    name      = "SSMParameterAccess"
    namespace = "${var.app_name}/${var.environment}/SSM"
    value     = "1"
    
    default_value = 0
  }
}

# CloudWatch Alarm for excessive parameter access
resource "aws_cloudwatch_metric_alarm" "ssm_parameter_access_high" {
  alarm_name          = "${var.app_name}-${var.environment}-ssm-parameter-access-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "SSMParameterAccess"
  namespace           = "${var.app_name}/${var.environment}/SSM"
  period              = "300"
  statistic           = "Sum"
  threshold           = var.ssm_access_threshold
  alarm_description   = "This metric monitors excessive SSM parameter access"
  alarm_actions       = var.alarm_actions

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-ssm-access-alarm"
    Component = "monitoring"
  })
}
