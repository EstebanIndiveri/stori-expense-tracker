# DynamoDB Module
module "dynamodb" {
  source = "./modules/dynamodb"
  
  environment  = var.environment
  project_name = var.project_name
}

# Lambda Functions Module
module "lambda" {
  source = "./modules/lambda"
  
  environment         = var.environment
  project_name        = var.project_name
  dynamodb_table_name = module.dynamodb.table_name
  dynamodb_table_arn  = module.dynamodb.table_arn
  openai_api_key_ssm  = module.ssm.openai_api_key_parameter_name
}

# API Gateway Module
module "api_gateway" {
  source = "./modules/api-gateway"
  
  environment            = var.environment
  project_name           = var.project_name
  cors_origins           = var.cors_origins
  api_lambda_function_arn = module.lambda.api_function_arn
  ai_lambda_function_arn  = module.lambda.ai_advisor_function_arn
  domain_name            = var.api_domain_name
  certificate_arn        = var.certificate_arn
}

# Frontend S3 + CloudFront Module
module "frontend" {
  source = "./modules/frontend"
  
  environment     = var.environment
  project_name    = var.project_name
  domain_name     = var.domain_name
  certificate_arn = var.certificate_arn
}

# SSM Parameters Module
module "ssm" {
  source = "./modules/ssm"
  
  environment    = var.environment
  project_name   = var.project_name
  openai_api_key = var.openai_api_key
}

# IAM Roles and Policies Module
module "iam" {
  source = "./modules/iam"
  
  environment         = var.environment
  project_name        = var.project_name
  dynamodb_table_arn  = module.dynamodb.table_arn
  ssm_parameter_arns  = [module.ssm.openai_api_key_parameter_arn]
}
