resource "aws_sqs_queue" "discovery_queue" {
  name                      = "terraform-discovery-queue"
  message_retention_seconds = 3600
  visibility_timeout_seconds = 100
  # redrive_policy = jsonencode({
  #   deadLetterTargetArn = aws_sqs_queue.terraform_queue_deadletter.arn
  #   maxReceiveCount     = 4
  # })
}

resource "aws_sqs_queue" "retrieval_queue" {
  name                      = "terraform-retrieval-queue"
  message_retention_seconds = 3600
  visibility_timeout_seconds = 100
  # redrive_policy = jsonencode({
  #   deadLetterTargetArn = aws_sqs_queue.terraform_queue_deadletter.arn
  #   maxReceiveCount     = 4
  # })
}

resource "aws_sqs_queue" "scan_queue" {
  name                      = "terraform-scan-queue"
  message_retention_seconds = 3600
  visibility_timeout_seconds = 100
  # redrive_policy = jsonencode({
  #   deadLetterTargetArn = aws_sqs_queue.terraform_queue_deadletter.arn
  #   maxReceiveCount     = 4
  # })
}