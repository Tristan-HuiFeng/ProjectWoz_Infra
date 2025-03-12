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
          "ses:SendRawEmail"
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
        "Resource": var.discovery_sqs_queue_arn
      }
    ]
  })
}

resource "aws_iam_policy_attachment" "custom_lambda_policy_attachment" {
  name       = "custom_lambda_policy_attachment"
  roles      = [aws_iam_role.lambda_role.name]
  policy_arn = aws_iam_policy.custom_lambda_policy.arn
}

resource "aws_security_group" "lambda" {
  vpc_id = "vpc-0f15d01f0801ba4a2"
  name   = "lambda_security_group"

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  # egress {
  #   protocol        = "tcp"
  #   from_port       = 443
  #   to_port         = 443
  #   cidr_blocks = ["0.0.0.0/0"]
  # }

  # egress {
  #   from_port   = 5000
  #   to_port     = 5000
  #   protocol    = "tcp"
  #   cidr_blocks = ["0.0.0.0/0"]
  # }

  # egress {
  #   from_port   = 80
  #   to_port     = 80
  #   protocol    = "tcp"
  #   cidr_blocks = ["0.0.0.0/0"]
  # }
}

##################################
# Discovery Lambda
##################################

data "archive_file" "discovery_lambda_zip" {
  type        = "zip"
  source_file = "${path.module}/../../../../bin/discovery/bootstrap"
  output_path = "${path.module}/../../../../bin/discovery/bootstrap.zip"
}

resource "aws_lambda_function" "discovery" {
  function_name = "notification_cs464_lambda"
  handler       = "notification_lambda.lambda_handler"
  runtime       = "provided.al2023"
  role          = aws_iam_role.lambda_role.arn

  filename = "${path.module}/../../../../bin/discovery/bootstrap.zip"

  vpc_config {
    subnet_ids         = ["subnet-04b7e2183fbe07ff9", "subnet-01df9b65cbec83278"]
    security_group_ids = [aws_security_group.lambda.id]
  }

  environment {
    variables = {
      MONGO_DB_STRING_PARAM = "/cs464/mongo_db_string"
    }
  }
  timeout          = 45
  source_code_hash = data.archive_file.discovery_lambda_zip.output_base64sha256
}

resource "aws_lambda_event_source_mapping" "sqs_event_source_mapping" {
  batch_size          = 1
  event_source_arn    = var.discovery_sqs_queue_arn
  function_name       = aws_lambda_function.discovery.function_name
  enabled             = true
}
