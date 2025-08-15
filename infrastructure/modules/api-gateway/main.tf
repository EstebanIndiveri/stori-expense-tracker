# HTTP API Gateway v2 for better performance and lower cost
resource "aws_apigatewayv2_api" "main" {
  name          = "${var.project_name}-api-${var.environment}"
  protocol_type = "HTTP"
  description   = "Stori Expense Tracker API - ${var.environment}"
  version       = "v1"
  
  # CORS configuration
  cors_configuration {
    allow_credentials = false
    allow_headers     = [
      "content-type",
      "x-amz-date",
      "authorization",
      "x-api-key",
      "x-amz-security-token",
      "x-amz-user-agent"
    ]
    allow_methods     = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_origins     = var.cors_origins
    expose_headers    = ["x-amz-request-id"]
    max_age          = 300
  }
  
  tags = merge(var.tags, {
    Name = "${var.project_name}-api-${var.environment}"
  })
}

# API Gateway Stage
resource "aws_apigatewayv2_stage" "main" {
  api_id      = aws_apigatewayv2_api.main.id
  name        = var.environment
  auto_deploy = true
  
  # Access logging
  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_gateway.arn
    format = jsonencode({
      requestId      = "$context.requestId"
      requestTime    = "$context.requestTime"
      requestTimeEpoch = "$context.requestTimeEpoch"
      httpMethod     = "$context.httpMethod"
      routeKey       = "$context.routeKey"
      status         = "$context.status"
      error          = "$context.error.message"
      responseLength = "$context.responseLength"
      ip             = "$context.identity.sourceIp"
      userAgent      = "$context.identity.userAgent"
      responseTime   = "$context.responseTime"
    })
  }
  
  # Default route settings
  default_route_settings {
    detailed_metrics_enabled = var.enable_detailed_metrics
    throttling_burst_limit   = var.throttling_burst_limit
    throttling_rate_limit    = var.throttling_rate_limit
  }
  
  tags = var.tags
}

# Lambda Integrations
resource "aws_apigatewayv2_integration" "api_lambda" {
  api_id             = aws_apigatewayv2_api.main.id
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
  integration_uri    = var.api_lambda_invoke_arn
  
  payload_format_version = "2.0"
  timeout_milliseconds   = 29000
  
  request_parameters = {
    "overwrite:header.x-request-id" = "$request.header.x-amz-request-id"
  }
}

resource "aws_apigatewayv2_integration" "ai_lambda" {
  api_id             = aws_apigatewayv2_api.main.id
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
  integration_uri    = var.ai_lambda_invoke_arn
  
  payload_format_version = "2.0"
  timeout_milliseconds   = 29000
}

# API Routes - Transactions
resource "aws_apigatewayv2_route" "get_transactions" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "GET /api/v1/transactions"
  target    = "integrations/${aws_apigatewayv2_integration.api_lambda.id}"
  
  authorization_type = var.enable_auth ? "JWT" : "NONE"
  authorizer_id      = var.enable_auth ? aws_apigatewayv2_authorizer.jwt[0].id : null
}

resource "aws_apigatewayv2_route" "create_transaction" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "POST /api/v1/transactions"
  target    = "integrations/${aws_apigatewayv2_integration.api_lambda.id}"
  
  authorization_type = var.enable_auth ? "JWT" : "NONE"
  authorizer_id      = var.enable_auth ? aws_apigatewayv2_authorizer.jwt[0].id : null
}

# API Routes - Analytics
resource "aws_apigatewayv2_route" "get_summary" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "GET /api/v1/analytics/summary"
  target    = "integrations/${aws_apigatewayv2_integration.api_lambda.id}"
  
  authorization_type = var.enable_auth ? "JWT" : "NONE"
  authorizer_id      = var.enable_auth ? aws_apigatewayv2_authorizer.jwt[0].id : null
}

resource "aws_apigatewayv2_route" "get_timeline" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "GET /api/v1/analytics/timeline"
  target    = "integrations/${aws_apigatewayv2_integration.api_lambda.id}"
  
  authorization_type = var.enable_auth ? "JWT" : "NONE"
  authorizer_id      = var.enable_auth ? aws_apigatewayv2_authorizer.jwt[0].id : null
}

resource "aws_apigatewayv2_route" "get_categories" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "GET /api/v1/analytics/categories"
  target    = "integrations/${aws_apigatewayv2_integration.api_lambda.id}"
  
  authorization_type = var.enable_auth ? "JWT" : "NONE"
  authorizer_id      = var.enable_auth ? aws_apigatewayv2_authorizer.jwt[0].id : null
}

# API Routes - AI Advisor
resource "aws_apigatewayv2_route" "get_ai_advice" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "POST /api/v1/ai/advice"
  target    = "integrations/${aws_apigatewayv2_integration.ai_lambda.id}"
  
  authorization_type = var.enable_auth ? "JWT" : "NONE"
  authorizer_id      = var.enable_auth ? aws_apigatewayv2_authorizer.jwt[0].id
}

# Catch-all route for handling unmatched requests
resource "aws_apigatewayv2_route" "catch_all" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "$default"
  target    = "integrations/${aws_apigatewayv2_integration.api_lambda.id}"
}

# JWT Authorizer (optional, for future auth implementation)
resource "aws_apigatewayv2_authorizer" "jwt" {
  count            = var.enable_auth ? 1 : 0
  api_id           = aws_apigatewayv2_api.main.id
  authorizer_type  = "JWT"
  identity_sources = ["$request.header.Authorization"]
  name             = "${var.project_name}-jwt-authorizer-${var.environment}"
  
  jwt_configuration {
    audience = var.jwt_audience
    issuer   = var.jwt_issuer
  }
}

# Custom Domain (optional)
resource "aws_apigatewayv2_domain_name" "api" {
  count       = var.domain_name != "" ? 1 : 0
  domain_name = var.domain_name
  
  domain_name_configuration {
    certificate_arn = var.certificate_arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }
  
  tags = var.tags
}

resource "aws_apigatewayv2_api_mapping" "api" {
  count       = var.domain_name != "" ? 1 : 0
  api_id      = aws_apigatewayv2_api.main.id
  domain_name = aws_apigatewayv2_domain_name.api[0].id
  stage       = aws_apigatewayv2_stage.main.id
}

# Route53 record for custom domain
resource "aws_route53_record" "api" {
  count   = var.domain_name != "" && var.hosted_zone_id != "" ? 1 : 0
  zone_id = var.hosted_zone_id
  name    = var.domain_name
  type    = "A"
  
  alias {
    name                   = aws_apigatewayv2_domain_name.api[0].domain_name_configuration[0].target_domain_name
    zone_id                = aws_apigatewayv2_domain_name.api[0].domain_name_configuration[0].hosted_zone_id
    evaluate_target_health = false
  }
}

# CloudWatch Log Group for API Gateway
resource "aws_cloudwatch_log_group" "api_gateway" {
  name              = "/aws/apigateway/${var.project_name}-${var.environment}"
  retention_in_days = var.log_retention_days
  kms_key_id        = var.kms_key_arn
  
  tags = var.tags
}

# API Gateway Account settings for CloudWatch
resource "aws_api_gateway_account" "main" {
  cloudwatch_role_arn = aws_iam_role.api_gateway_cloudwatch.arn
}

# IAM Role for API Gateway CloudWatch logging
resource "aws_iam_role" "api_gateway_cloudwatch" {
  name = "${var.project_name}-api-gateway-cloudwatch-${var.environment}"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "apigateway.amazonaws.com"
        }
      }
    ]
  })
  
  tags = var.tags
}

resource "aws_iam_role_policy_attachment" "api_gateway_cloudwatch" {
  role       = aws_iam_role.api_gateway_cloudwatch.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonAPIGatewayPushToCloudWatchLogs"
}

# CloudWatch Alarms
resource "aws_cloudwatch_metric_alarm" "api_gateway_4xx_errors" {
  alarm_name          = "${var.project_name}-api-gateway-4xx-errors-${var.environment}"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "4XXError"
  namespace           = "AWS/ApiGatewayV2"
  period              = "300"
  statistic           = "Sum"
  threshold           = "10"
  alarm_description   = "This metric monitors API Gateway 4XX errors"
  alarm_actions       = var.sns_alarm_topic_arn != "" ? [var.sns_alarm_topic_arn] : []
  
  dimensions = {
    ApiId = aws_apigatewayv2_api.main.id
    Stage = aws_apigatewayv2_stage.main.name
  }
  
  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "api_gateway_5xx_errors" {
  alarm_name          = "${var.project_name}-api-gateway-5xx-errors-${var.environment}"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "5XXError"
  namespace           = "AWS/ApiGatewayV2"
  period              = "300"
  statistic           = "Sum"
  threshold           = "5"
  alarm_description   = "This metric monitors API Gateway 5XX errors"
  alarm_actions       = var.sns_alarm_topic_arn != "" ? [var.sns_alarm_topic_arn] : []
  
  dimensions = {
    ApiId = aws_apigatewayv2_api.main.id
    Stage = aws_apigatewayv2_stage.main.name
  }
  
  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "api_gateway_latency" {
  alarm_name          = "${var.project_name}-api-gateway-latency-${var.environment}"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "IntegrationLatency"
  namespace           = "AWS/ApiGatewayV2"
  period              = "300"
  statistic           = "Average"
  threshold           = "10000"  # 10 seconds
  alarm_description   = "This metric monitors API Gateway integration latency"
  alarm_actions       = var.sns_alarm_topic_arn != "" ? [var.sns_alarm_topic_arn] : []
  
  dimensions = {
    ApiId = aws_apigatewayv2_api.main.id
    Stage = aws_apigatewayv2_stage.main.name
  }
  
  tags = var.tags
}
