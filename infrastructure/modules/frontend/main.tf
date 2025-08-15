# S3 bucket for hosting static website
resource "aws_s3_bucket" "frontend" {
  bucket        = "${var.app_name}-${var.environment}-frontend-${random_string.bucket_suffix.result}"
  force_destroy = var.environment != "prod"

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-frontend"
    Component = "frontend"
    Service   = "s3"
  })
}

# Random suffix for bucket name uniqueness
resource "random_string" "bucket_suffix" {
  length  = 8
  special = false
  upper   = false
}

# S3 bucket versioning
resource "aws_s3_bucket_versioning" "frontend" {
  bucket = aws_s3_bucket.frontend.id
  versioning_configuration {
    status = "Enabled"
  }
}

# S3 bucket encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "frontend" {
  bucket = aws_s3_bucket.frontend.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = var.kms_key_id
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

# S3 bucket public access block
resource "aws_s3_bucket_public_access_block" "frontend" {
  bucket = aws_s3_bucket.frontend.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# S3 bucket policy for CloudFront access
resource "aws_s3_bucket_policy" "frontend" {
  bucket = aws_s3_bucket.frontend.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowCloudFrontServicePrincipal"
        Effect = "Allow"
        Principal = {
          Service = "cloudfront.amazonaws.com"
        }
        Action   = "s3:GetObject"
        Resource = "${aws_s3_bucket.frontend.arn}/*"
        Condition = {
          StringEquals = {
            "AWS:SourceArn" = aws_cloudfront_distribution.frontend.arn
          }
        }
      }
    ]
  })

  depends_on = [aws_cloudfront_distribution.frontend]
}

# CloudFront Origin Access Control
resource "aws_cloudfront_origin_access_control" "frontend" {
  name                              = "${var.app_name}-${var.environment}-frontend-oac"
  description                       = "OAC for ${var.app_name} ${var.environment} frontend"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

# CloudFront distribution
resource "aws_cloudfront_distribution" "frontend" {
  enabled             = true
  is_ipv6_enabled     = true
  default_root_object = "index.html"
  price_class         = var.cloudfront_price_class
  aliases             = var.domain_name != "" ? [var.domain_name] : []

  origin {
    domain_name              = aws_s3_bucket.frontend.bucket_regional_domain_name
    origin_id                = "S3-${aws_s3_bucket.frontend.id}"
    origin_access_control_id = aws_cloudfront_origin_access_control.frontend.id
  }

  default_cache_behavior {
    allowed_methods        = ["GET", "HEAD", "OPTIONS"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "S3-${aws_s3_bucket.frontend.id}"
    compress               = true
    viewer_protocol_policy = "redirect-to-https"

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }

    min_ttl     = 0
    default_ttl = var.cache_default_ttl
    max_ttl     = var.cache_max_ttl

    response_headers_policy_id = aws_cloudfront_response_headers_policy.frontend.id
  }

  # Cache behavior for static assets
  ordered_cache_behavior {
    path_pattern           = "/static/*"
    allowed_methods        = ["GET", "HEAD"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "S3-${aws_s3_bucket.frontend.id}"
    compress               = true
    viewer_protocol_policy = "redirect-to-https"

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }

    min_ttl     = 86400  # 1 day
    default_ttl = 604800 # 1 week
    max_ttl     = 31536000 # 1 year
  }

  # Custom error pages for SPA routing
  custom_error_response {
    error_code         = 403
    response_code      = 200
    response_page_path = "/index.html"
    error_caching_min_ttl = 0
  }

  custom_error_response {
    error_code         = 404
    response_code      = 200
    response_page_path = "/index.html"
    error_caching_min_ttl = 0
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = var.domain_name == ""
    acm_certificate_arn            = var.domain_name != "" ? var.acm_certificate_arn : null
    ssl_support_method             = var.domain_name != "" ? "sni-only" : null
    minimum_protocol_version       = var.domain_name != "" ? "TLSv1.2_2021" : null
  }

  web_acl_id = var.waf_web_acl_arn != "" ? var.waf_web_acl_arn : null

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-frontend-distribution"
    Component = "frontend"
    Service   = "cloudfront"
  })
}

# CloudFront Response Headers Policy for security
resource "aws_cloudfront_response_headers_policy" "frontend" {
  name    = "${var.app_name}-${var.environment}-frontend-security-headers"
  comment = "Security headers policy for ${var.app_name} frontend"

  security_headers_config {
    strict_transport_security {
      access_control_max_age_sec = 31536000
      include_subdomains         = true
      override                   = false
    }

    content_type_options {
      override = false
    }

    frame_options {
      frame_option = "DENY"
      override     = false
    }

    referrer_policy {
      referrer_policy = "strict-origin-when-cross-origin"
      override        = false
    }
  }

  cors_config {
    access_control_allow_credentials = false
    access_control_allow_headers {
      items = ["*"]
    }
    access_control_allow_methods {
      items = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    }
    access_control_allow_origins {
      items = var.cors_allowed_origins
    }
    access_control_max_age_sec = 86400
    origin_override           = false
  }
}

# CloudWatch Log Group for CloudFront
resource "aws_cloudwatch_log_group" "cloudfront" {
  count             = var.enable_cloudfront_logging ? 1 : 0
  name              = "/aws/cloudfront/${var.app_name}/${var.environment}/frontend"
  retention_in_days = var.log_retention_days
  kms_key_id        = var.kms_key_id

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-cloudfront-logs"
    Component = "monitoring"
    Service   = "cloudfront"
  })
}

# CloudWatch Alarms for monitoring
resource "aws_cloudwatch_metric_alarm" "cloudfront_error_rate" {
  alarm_name          = "${var.app_name}-${var.environment}-cloudfront-error-rate"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "ErrorRate"
  namespace           = "AWS/CloudFront"
  period              = "300"
  statistic           = "Average"
  threshold           = var.error_rate_threshold
  alarm_description   = "This metric monitors CloudFront error rate"
  alarm_actions       = var.alarm_actions

  dimensions = {
    DistributionId = aws_cloudfront_distribution.frontend.id
  }

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-cloudfront-error-rate-alarm"
    Component = "monitoring"
    Service   = "cloudfront"
  })
}

resource "aws_cloudfront_monitoring_subscription" "frontend" {
  count                   = var.enable_real_time_metrics ? 1 : 0
  distribution_id         = aws_cloudfront_distribution.frontend.id

  monitoring_subscription {
    realtime_metrics_subscription_config {
      realtime_metrics_subscription_status = "Enabled"
    }
  }
}

# Route53 record (if custom domain is used)
resource "aws_route53_record" "frontend" {
  count   = var.domain_name != "" && var.route53_zone_id != "" ? 1 : 0
  zone_id = var.route53_zone_id
  name    = var.domain_name
  type    = "A"

  alias {
    name                   = aws_cloudfront_distribution.frontend.domain_name
    zone_id                = aws_cloudfront_distribution.frontend.hosted_zone_id
    evaluate_target_health = false
  }
}

# S3 bucket for deployment artifacts/builds
resource "aws_s3_bucket" "build_artifacts" {
  bucket        = "${var.app_name}-${var.environment}-build-artifacts-${random_string.artifacts_suffix.result}"
  force_destroy = true

  tags = merge(var.tags, {
    Name      = "${var.app_name}-${var.environment}-build-artifacts"
    Component = "ci-cd"
    Service   = "s3"
  })
}

resource "random_string" "artifacts_suffix" {
  length  = 8
  special = false
  upper   = false
}

resource "aws_s3_bucket_versioning" "build_artifacts" {
  bucket = aws_s3_bucket.build_artifacts.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "build_artifacts" {
  bucket = aws_s3_bucket.build_artifacts.id

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = var.kms_key_id
      sse_algorithm     = "aws:kms"
    }
    bucket_key_enabled = true
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "build_artifacts" {
  bucket = aws_s3_bucket.build_artifacts.id

  rule {
    id     = "delete_old_builds"
    status = "Enabled"

    expiration {
      days = 30
    }

    noncurrent_version_expiration {
      noncurrent_days = 7
    }
  }
}
