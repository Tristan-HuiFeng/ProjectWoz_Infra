package s3

# Check if the bucket has public access (should deny any access for *)
deny_public_access {
    input.Statement[_].Principal == "*"
    input.Statement[_].Effect == "Allow"
    input.Statement[_].Action == "s3:*"
    input.Statement[_].Resource == "arn:aws:s3:::*"
}

# Check if the bucket has unrestricted access (e.g., * is used for actions or resources)
deny_unrestricted_access {
    input.Statement[_].Principal == "*"
    input.Statement[_].Effect == "Allow"
    input.Statement[_].Action == "s3:*"
    input.Statement[_].Resource == "arn:aws:s3:::my-bucket/*"
}

# Check if the bucket has no encryption (should have encryption enabled)
deny_missing_encryption {
    not input.Statement[_].Condition.encryption
}

# Ensure only specific IPs have access (whitelisted IPs)
deny_non_whitelisted_ip {
    not input.Statement[_].Condition.IpAddress["aws:SourceIp"] == "10.0.0.0/24"
}