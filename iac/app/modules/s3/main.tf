resource "aws_s3_bucket" "lambda" {
  bucket = "cs464-lambda-s3"

  tags = {
    Environment = "Prod"
  }
}

resource "aws_s3_bucket_versioning" "lambda_versioning" {
  bucket = aws_s3_bucket.lambda.id
  versioning_configuration {
    status = "Enabled"
  }
}