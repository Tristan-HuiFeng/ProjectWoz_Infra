output "discovery_sqs_queue_arn" {
  value = aws_sqs_queue.discovery_queue.arn
}

output "discovery_sqs_queue_url" {
  value = aws_sqs_queue.discovery_queue.url
}

output "retrieval_sqs_queue_arn" {
  value = aws_sqs_queue.retrieval_queue.arn
}

output "retrieval_sqs_queue_url" {
  value = aws_sqs_queue.retrieval_queue.url
}