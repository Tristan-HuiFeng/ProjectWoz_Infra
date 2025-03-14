package s3

default allow := false

deny[msg] if {
    bucket_policy := input.bucket_policy
    statement := bucket_policy.Statement[_]
    statement.Principal == "*"
    msg := "Principal is too wide. Restrict access to specific roles or users."
}

deny[msg] if {
    bucket_policy := input.bucket_policy
    statement := bucket_policy.Statement[_]
    statement.Condition["IpAddress"]["aws:SourceIp"] == "0.0.0.0/0"
    msg := "The SourceIp condition allows access from any IP address. Restrict access to specific IP ranges."
}

deny[msg] if {
    bucket_policy := input.bucket_policy
    statement := bucket_policy.Statement[_]
    statement.Principal == "*"
    not statement.Condition
    msg := "Allowing access to all (Principal: *) without conditions is a security risk. Please add restrictive conditions."
}

allow if {
    count(deny) == 0
}