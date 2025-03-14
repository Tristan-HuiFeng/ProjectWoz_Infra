variable "discovery_sqs_queue_arn" {
  description = "The arn for discovery sqs"
  type        = string
}

variable "retrieval_sqs_queue_arn" {
  description = "The arn for retrieval sqs"
  type        = string
}

variable "scan_sqs_queue_arn" {
  description = "The arn for scan sqs"
  type        = string
}
