output "s3_bucket_name" {
  description = "Name of the S3 bucket for frontend hosting"
  value       = aws_s3_bucket.frontend.bucket
}

output "s3_bucket_arn" {
  description = "ARN of the S3 bucket for frontend hosting"
  value       = aws_s3_bucket.frontend.arn
}

output "s3_bucket_domain_name" {
  description = "Domain name of the S3 bucket"
  value       = aws_s3_bucket.frontend.bucket_domain_name
}

output "s3_bucket_regional_domain_name" {
  description = "Regional domain name of the S3 bucket"
  value       = aws_s3_bucket.frontend.bucket_regional_domain_name
}

output "cloudfront_distribution_id" {
  description = "ID of the CloudFront distribution"
  value       = aws_cloudfront_distribution.frontend.id
}

output "cloudfront_distribution_arn" {
  description = "ARN of the CloudFront distribution"
  value       = aws_cloudfront_distribution.frontend.arn
}

output "cloudfront_domain_name" {
  description = "Domain name of the CloudFront distribution"
  value       = aws_cloudfront_distribution.frontend.domain_name
}

output "cloudfront_hosted_zone_id" {
  description = "Hosted zone ID of the CloudFront distribution"
  value       = aws_cloudfront_distribution.frontend.hosted_zone_id
}

output "frontend_url" {
  description = "URL of the frontend application"
  value       = var.domain_name != "" ? "https://${var.domain_name}" : "https://${aws_cloudfront_distribution.frontend.domain_name}"
}

output "custom_domain_name" {
  description = "Custom domain name (if configured)"
  value       = var.domain_name
}

output "build_artifacts_bucket_name" {
  description = "Name of the S3 bucket for build artifacts"
  value       = aws_s3_bucket.build_artifacts.bucket
}

output "build_artifacts_bucket_arn" {
  description = "ARN of the S3 bucket for build artifacts"
  value       = aws_s3_bucket.build_artifacts.arn
}

output "origin_access_control_id" {
  description = "ID of the CloudFront Origin Access Control"
  value       = aws_cloudfront_origin_access_control.frontend.id
}

output "response_headers_policy_id" {
  description = "ID of the CloudFront Response Headers Policy"
  value       = aws_cloudfront_response_headers_policy.frontend.id
}

output "cloudfront_log_group_name" {
  description = "Name of the CloudFront log group"
  value       = var.enable_cloudfront_logging ? aws_cloudwatch_log_group.cloudfront[0].name : ""
}

output "route53_record_name" {
  description = "Name of the Route53 record (if created)"
  value       = var.domain_name != "" && var.route53_zone_id != "" ? aws_route53_record.frontend[0].name : ""
}

output "deployment_info" {
  description = "Deployment information for CI/CD"
  value = {
    s3_bucket                = aws_s3_bucket.frontend.bucket
    cloudfront_distribution  = aws_cloudfront_distribution.frontend.id
    build_artifacts_bucket   = aws_s3_bucket.build_artifacts.bucket
    frontend_url            = var.domain_name != "" ? "https://${var.domain_name}" : "https://${aws_cloudfront_distribution.frontend.domain_name}"
  }
}
