# Project Woz

## 🚀 Prerequisikte

- MongoDB
- Setting up SSM Parameters in AWS
- Terraform
- AWS CLI with profile configured
- Change the AWS Account ID and SSM parameters name to fit your configuration under iac/app

## 🛠️ Installation
```
cd iac/app
terraform plan -o tfplan
terraform apply tfplan
```

## GCP Remediation
This bash file resolves a small subset of issues like lack of public access prevention and soft delete policy. It is meant as a POC.
1. Copy the bash file in CloudShell editor.
2. Run ./gcp_remediation.sh in the CloudShell terminal.
