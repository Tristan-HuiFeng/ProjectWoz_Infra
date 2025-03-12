terraform {
  required_version = ">= 1.9.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">=5.90.0"
    }
  }

  backend "s3" {
    bucket         = "cs464-terraform-state"
    key            = "state/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "cs464-terraform-state-table"
  }
}

provider "aws" {
  region  = "us-east-1"
  // profile = "wozrole"
}
