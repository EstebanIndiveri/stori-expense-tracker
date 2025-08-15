# Lambda execution role for API functions
resource "aws_iam_role" "lambda_api_execution" {
  name = "${var.app_name}-${var.environment}-lambda-api-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-lambda-api-execution"
    Component = "security"
    Service   = "lambda"
  })
}

# Lambda execution role for AI advisor functions
resource "aws_iam_role" "lambda_ai_execution" {
  name = "${var.app_name}-${var.environment}-lambda-ai-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-lambda-ai-execution"
    Component = "security"
    Service   = "lambda"
  })
}

# Basic Lambda execution policy attachment
resource "aws_iam_role_policy_attachment" "lambda_api_basic" {
  role       = aws_iam_role.lambda_api_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "lambda_ai_basic" {
  role       = aws_iam_role.lambda_ai_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# VPC access policy attachment (if using VPC)
resource "aws_iam_role_policy_attachment" "lambda_api_vpc" {
  count      = var.enable_vpc_access ? 1 : 0
  role       = aws_iam_role.lambda_api_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

resource "aws_iam_role_policy_attachment" "lambda_ai_vpc" {
  count      = var.enable_vpc_access ? 1 : 0
  role       = aws_iam_role.lambda_ai_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

# X-Ray tracing policy attachment
resource "aws_iam_role_policy_attachment" "lambda_api_xray" {
  role       = aws_iam_role.lambda_api_execution.name
  policy_arn = "arn:aws:iam::aws:policy/AWSXRayDaemonWriteAccess"
}

resource "aws_iam_role_policy_attachment" "lambda_ai_xray" {
  role       = aws_iam_role.lambda_ai_execution.name
  policy_arn = "arn:aws:iam::aws:policy/AWSXRayDaemonWriteAccess"
}

# DynamoDB access policy for API Lambda
resource "aws_iam_policy" "lambda_api_dynamodb" {
  name        = "${var.app_name}-${var.environment}-lambda-api-dynamodb"
  description = "DynamoDB access policy for API Lambda functions"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Query",
          "dynamodb:Scan",
          "dynamodb:BatchGetItem",
          "dynamodb:BatchWriteItem"
        ]
        Resource = [
          var.dynamodb_table_arn,
          "${var.dynamodb_table_arn}/index/*"
        ]
      }
    ]
  })

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-lambda-api-dynamodb"
    Component = "security"
    Service   = "dynamodb"
  })
}

resource "aws_iam_role_policy_attachment" "lambda_api_dynamodb" {
  role       = aws_iam_role.lambda_api_execution.name
  policy_arn = aws_iam_policy.lambda_api_dynamodb.arn
}

# SSM parameter access policy for both Lambda functions
resource "aws_iam_policy" "lambda_ssm_parameters" {
  name        = "${var.app_name}-${var.environment}-lambda-ssm-parameters"
  description = "SSM parameter access policy for Lambda functions"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ssm:GetParameter",
          "ssm:GetParameters",
          "ssm:GetParametersByPath"
        ]
        Resource = [
          "arn:aws:ssm:${var.aws_region}:*:parameter/${var.app_name}/${var.environment}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "kms:Decrypt"
        ]
        Resource = [
          var.kms_key_arn
        ]
        Condition = {
          StringEquals = {
            "kms:ViaService" = "ssm.${var.aws_region}.amazonaws.com"
          }
        }
      }
    ]
  })

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-lambda-ssm-parameters"
    Component = "security"
    Service   = "ssm"
  })
}

resource "aws_iam_role_policy_attachment" "lambda_api_ssm" {
  role       = aws_iam_role.lambda_api_execution.name
  policy_arn = aws_iam_policy.lambda_ssm_parameters.arn
}

resource "aws_iam_role_policy_attachment" "lambda_ai_ssm" {
  role       = aws_iam_role.lambda_ai_execution.name
  policy_arn = aws_iam_policy.lambda_ssm_parameters.arn
}

# CloudWatch Logs policy for structured logging
resource "aws_iam_policy" "lambda_cloudwatch_logs" {
  name        = "${var.app_name}-${var.environment}-lambda-cloudwatch-logs"
  description = "Enhanced CloudWatch Logs access for Lambda functions"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "logs:DescribeLogGroups",
          "logs:DescribeLogStreams"
        ]
        Resource = [
          "arn:aws:logs:${var.aws_region}:*:log-group:/aws/lambda/${var.app_name}-${var.environment}-*",
          "arn:aws:logs:${var.aws_region}:*:log-group:/aws/lambda/${var.app_name}-${var.environment}-*:*"
        ]
      }
    ]
  })

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-lambda-cloudwatch-logs"
    Component = "security"
    Service   = "cloudwatch"
  })
}

resource "aws_iam_role_policy_attachment" "lambda_api_cloudwatch" {
  role       = aws_iam_role.lambda_api_execution.name
  policy_arn = aws_iam_policy.lambda_cloudwatch_logs.arn
}

resource "aws_iam_role_policy_attachment" "lambda_ai_cloudwatch" {
  role       = aws_iam_role.lambda_ai_execution.name
  policy_arn = aws_iam_policy.lambda_cloudwatch_logs.arn
}

# DynamoDB read-only access for AI advisor (for analytics)
resource "aws_iam_policy" "lambda_ai_dynamodb_read" {
  name        = "${var.app_name}-${var.environment}-lambda-ai-dynamodb-read"
  description = "DynamoDB read-only access for AI advisor Lambda"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:Query",
          "dynamodb:Scan",
          "dynamodb:BatchGetItem"
        ]
        Resource = [
          var.dynamodb_table_arn,
          "${var.dynamodb_table_arn}/index/*"
        ]
      }
    ]
  })

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-lambda-ai-dynamodb-read"
    Component = "security"
    Service   = "dynamodb"
  })
}

resource "aws_iam_role_policy_attachment" "lambda_ai_dynamodb_read" {
  role       = aws_iam_role.lambda_ai_execution.name
  policy_arn = aws_iam_policy.lambda_ai_dynamodb_read.arn
}

# API Gateway logging role
resource "aws_iam_role" "api_gateway_logging" {
  name = "${var.app_name}-${var.environment}-api-gateway-logging"

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

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-api-gateway-logging"
    Component = "security"
    Service   = "api-gateway"
  })
}

resource "aws_iam_role_policy_attachment" "api_gateway_logging" {
  role       = aws_iam_role.api_gateway_logging.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonAPIGatewayPushToCloudWatchLogs"
}

# CloudFront Origin Access Control (if using S3 for frontend)
resource "aws_iam_policy" "cloudfront_s3_access" {
  count       = var.enable_frontend_deployment ? 1 : 0
  name        = "${var.app_name}-${var.environment}-cloudfront-s3-access"
  description = "CloudFront access to S3 bucket for frontend"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject"
        ]
        Resource = [
          "${var.frontend_s3_bucket_arn}/*"
        ]
      }
    ]
  })

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-cloudfront-s3-access"
    Component = "security"
    Service   = "cloudfront"
  })
}

# GitHub Actions deployment role (for CI/CD)
resource "aws_iam_role" "github_actions_deployment" {
  count = var.enable_github_actions_role ? 1 : 0
  name  = "${var.app_name}-${var.environment}-github-actions-deployment"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = var.github_actions_oidc_provider_arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          }
          StringLike = {
            "token.actions.githubusercontent.com:sub" = "repo:${var.github_repository}:*"
          }
        }
      }
    ]
  })

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-github-actions-deployment"
    Component = "security"
    Service   = "github-actions"
  })
}

# GitHub Actions deployment policy
resource "aws_iam_policy" "github_actions_deployment" {
  count       = var.enable_github_actions_role ? 1 : 0
  name        = "${var.app_name}-${var.environment}-github-actions-deployment"
  description = "Deployment permissions for GitHub Actions"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "lambda:UpdateFunctionCode",
          "lambda:UpdateFunctionConfiguration",
          "lambda:GetFunction",
          "lambda:PublishVersion",
          "lambda:CreateAlias",
          "lambda:UpdateAlias"
        ]
        Resource = [
          "arn:aws:lambda:${var.aws_region}:*:function:${var.app_name}-${var.environment}-*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject",
          "s3:GetObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = var.enable_frontend_deployment ? [
          var.frontend_s3_bucket_arn,
          "${var.frontend_s3_bucket_arn}/*"
        ] : []
      },
      {
        Effect = "Allow"
        Action = [
          "cloudfront:CreateInvalidation"
        ]
        Resource = var.enable_frontend_deployment ? [
          var.cloudfront_distribution_arn
        ] : []
      }
    ]
  })

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-github-actions-deployment"
    Component = "security"
    Service   = "github-actions"
  })
}

resource "aws_iam_role_policy_attachment" "github_actions_deployment" {
  count      = var.enable_github_actions_role ? 1 : 0
  role       = aws_iam_role.github_actions_deployment[0].name
  policy_arn = aws_iam_policy.github_actions_deployment[0].arn
}
