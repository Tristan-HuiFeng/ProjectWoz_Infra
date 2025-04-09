resource "aws_iam_role" "lambda_role" {
  name = "CS464_Lambda_Function_Role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
        Effect = "Allow"
        Sid    = ""
      }
    ]
  })
}

resource "aws_iam_policy" "custom_lambda_policy" {
  name        = "custom_lambda_policy"
  description = "Custom permissions for Lambda function to access other services"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow",
        Action = [
          "secretsmanager:GetSecretValue",
          "ec2:CreateNetworkInterface",
          "ec2:DeleteNetworkInterface",
          "ec2:DescribeNetworkInterfaces",
          "ec2:DescribeSecurityGroups",
          "ec2:DescribeSubnets",
          "ec2:DescribeVpcs",
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "ses:SendEmail",
          "ses:SendRawEmail",
          "ssm:GetParameters",
          "ssm:GetParameter",
          "ssm:DescribeParameters"
        ],
        Resource = "*"
      },
      {
        "Effect" : "Allow",
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
        ],
        "Resource": [var.retrieval_sqs_queue_arn, var.scan_sqs_queue_arn]
      },
            {
        "Effect" : "Allow",
        Action = [
          "sqs:SendMessage",
          "sqs:GetQueueAttributes",
        ],
        "Resource": [var.retrieval_sqs_queue_arn, var.scan_sqs_queue_arn]
      }
    ]
  })
}

resource "aws_iam_policy_attachment" "custom_lambda_policy_attachment" {
  name       = "custom_lambda_policy_attachment"
  roles      = [aws_iam_role.lambda_role.name]
  policy_arn = aws_iam_policy.custom_lambda_policy.arn
}

# resource "aws_security_group" "lambda" {
#   vpc_id = "vpc-0f15d01f0801ba4a2"
#   name   = "lambda_security_group"

#   egress {
#     protocol    = "-1"
#     from_port   = 0
#     to_port     = 0
#     cidr_blocks = ["0.0.0.0/0"]
#   }

#   # egress {
#   #   protocol        = "tcp"
#   #   from_port       = 443
#   #   to_port         = 443
#   #   cidr_blocks = ["0.0.0.0/0"]
#   # }

#   # egress {
#   #   from_port   = 5000
#   #   to_port     = 5000
#   #   protocol    = "tcp"
#   #   cidr_blocks = ["0.0.0.0/0"]
#   # }

#   # egress {
#   #   from_port   = 80
#   #   to_port     = 80
#   #   protocol    = "tcp"
#   #   cidr_blocks = ["0.0.0.0/0"]
#   # }
# }

##################################
# Discovery Lambda
##################################

data "archive_file" "discovery_lambda_zip" {
  type        = "zip"
  # source_files = [
  #   "${path.module}/../../../../bin/discovery/bootstrap", 
  #   "${path.module}/../../../../bin/discovery/clientLibraryConfig-awswoz.json"
  # ]
  source_dir = "${path.module}/../../../../bin/discovery/"
  output_path = "${path.module}/../../../../bin/discovery/bootstrap.zip"
}

resource "aws_lambda_function" "discovery" {
  function_name = "discovery_cs464_lambda"
  handler       = "discovery_lambda.handler"
  runtime       = "provided.al2023"
  role          = aws_iam_role.lambda_role.arn

  filename = "${path.module}/../../../../bin/discovery/bootstrap.zip"

  # vpc_config {
  #   subnet_ids         = ["subnet-04b7e2183fbe07ff9", "subnet-01df9b65cbec83278"]
  #   security_group_ids = [aws_security_group.lambda.id]
  # }

  environment {
    variables = {
      MONGO_DB_STRING_PARAM = "/cs464/mongo_db_string"
      PROCESSING_ROLE = "/cs464/cross_account_role"
      RETRIEVAL_QUEUE_PARAM = "/cs464/retrieval_queue_url"
      GOOGLE_APPLICATION_CREDENTIALS = "clientLibraryConfig-awswoz.json"
      GOOGLE_CLOUD_PROJECT = "cs464-454011"
    }
  }
  timeout          = 45
  source_code_hash = data.archive_file.discovery_lambda_zip.output_base64sha256

  publish = true
}

resource "aws_lambda_alias" "discovery_alias" {
  name             = "live"
  function_name    = aws_lambda_function.discovery.function_name
  function_version = aws_lambda_function.discovery.version
}

# resource "aws_lambda_event_source_mapping" "discovery" {
#   batch_size          = 1
#   event_source_arn    = var.discovery_sqs_queue_arn
#   function_name       = aws_lambda_function.discovery.function_name
#   enabled             = true
# }

##################################
# Retrieval Lambda
##################################

data "archive_file" "retrieval_lambda_zip" {
  type        = "zip"
  source_dir = "${path.module}/../../../../bin/retrieval/"
  output_path = "${path.module}/../../../../bin/retrieval/bootstrap.zip"
}

resource "aws_lambda_function" "retrieval" {
  function_name = "retrieval_cs464_lambda"
  handler       = "retrieval_lambda.handler"
  runtime       = "provided.al2023"
  role          = aws_iam_role.lambda_role.arn

  filename = "${path.module}/../../../../bin/retrieval/bootstrap.zip"

  # vpc_config {
  #   subnet_ids         = ["subnet-04b7e2183fbe07ff9", "subnet-01df9b65cbec83278"]
  #   security_group_ids = [aws_security_group.lambda.id]
  # }

  environment {
    variables = {
      MONGO_DB_STRING_PARAM = "/cs464/mongo_db_string"
      PROCESSING_ROLE = "/cs464/cross_account_role"
      SCAN_QUEUE_PARAM = "/cs464/scan_queue_url"
    }
  }
  timeout          = 45
  source_code_hash = data.archive_file.retrieval_lambda_zip.output_base64sha256

  publish = true
}

resource "aws_lambda_alias" "retrieval_alias" {
  name             = "live"
  function_name    = aws_lambda_function.retrieval.function_name
  function_version = aws_lambda_function.retrieval.version
}

resource "aws_lambda_event_source_mapping" "retrieval" {
  batch_size          = 1
  event_source_arn    = var.retrieval_sqs_queue_arn
  function_name       = aws_lambda_function.retrieval.function_name
  enabled             = true
}

##################################
# Scan Lambda
##################################

data "archive_file" "scan_lambda_zip" {
  type        = "zip"
  source_file = "${path.module}/../../../../bin/scan/bootstrap"
  output_path = "${path.module}/../../../../bin/scan/bootstrap.zip"
}

resource "aws_lambda_function" "scan" {
  function_name = "scan_cs464_lambda"
  handler       = "scan_lambda.handler"
  runtime       = "provided.al2023"
  role          = aws_iam_role.lambda_role.arn

  filename = "${path.module}/../../../../bin/scan/bootstrap.zip"

  # vpc_config {
  #   subnet_ids         = ["subnet-04b7e2183fbe07ff9", "subnet-01df9b65cbec83278"]
  #   security_group_ids = [aws_security_group.lambda.id]
  # }

  environment {
    variables = {
      MONGO_DB_STRING_PARAM = "/cs464/mongo_db_string"
      PROCESSING_ROLE = "/cs464/cross_account_role"
      SMTP_PASSWORD_PARAM = "/cs464/smtp_password"
      SMTP_HOST="smtp.gmail.com"
      SMTP_PORT="587"
      SMTP_USER="flyingduckservices@gmail.com"
    }
  }
  timeout          = 45
  source_code_hash = data.archive_file.scan_lambda_zip.output_base64sha256

  publish = true
  
}

resource "aws_lambda_alias" "scan_alias" {
  name             = "live"
  function_name    = aws_lambda_function.scan.function_name
  function_version = aws_lambda_function.scan.version
}

resource "aws_lambda_event_source_mapping" "scan" {
  batch_size          = 1
  event_source_arn    = var.scan_sqs_queue_arn
  function_name       = aws_lambda_function.scan.function_name
  enabled             = true
}
