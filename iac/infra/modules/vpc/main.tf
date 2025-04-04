resource "aws_vpc" "main" {
  cidr_block           = "10.16.0.0/16" # VPC CIDR block
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = "main-cs464"
  }
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "main-igw"
  }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name = "public-route-table"
  }
}

resource "aws_route_table_association" "public_subnet_web" {
  subnet_id      = aws_subnet.web.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "public_subnet_app" {
  subnet_id      = aws_subnet.app.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "public_subnet_db" {
  subnet_id      = aws_subnet.db.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "public_subnet_web_2" {
  subnet_id      = aws_subnet.web_2.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "public_subnet_app_2" {
  subnet_id      = aws_subnet.app_2.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "public_subnet_db_2" {
  subnet_id      = aws_subnet.db_2.id
  route_table_id = aws_route_table.public.id
}

resource "aws_subnet" "web" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.16.0.0/20" # Subnet A
  availability_zone = "us-east-1a"

  tags = {
    Name = "web"
  }
}

resource "aws_subnet" "app" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.16.16.0/20" # Subnet B
  availability_zone = "us-east-1a"

  tags = {
    Name = "app"
  }
}

resource "aws_subnet" "db" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.16.32.0/20" # Subnet C
  availability_zone = "us-east-1a"

  tags = {
    Name = "db"
  }
}

resource "aws_subnet" "other" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.16.48.0/20" # Subnet D
  availability_zone = "us-east-1a"

  tags = {
    Name = "other"
  }
}

resource "aws_subnet" "web_2" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.16.64.0/20" # Subnet E
  availability_zone = "us-east-1b"

  tags = {
    Name = "web-2"
  }
}

resource "aws_subnet" "app_2" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.16.80.0/20" # Subnet F
  availability_zone = "us-east-1b"

  tags = {
    Name = "app-2"
  }
}

resource "aws_subnet" "db_2" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.16.96.0/20" # Subnet G
  availability_zone = "us-east-1b"

  tags = {
    Name = "db-2"
  }
}

resource "aws_subnet" "other_2" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.16.112.0/20" # Subnet H
  availability_zone = "us-east-1b"

  tags = {
    Name = "other-2"
  }
}

resource "aws_eip" "main" {
  domain = "vpc"
}


# resource "aws_nat_gateway" "main" {
#   allocation_id = aws_eip.main.id
#   subnet_id     = aws_subnet.web.id

#   # To ensure proper ordering, it is recommended to add an explicit dependency
#   # on the Internet Gateway for the VPC.
#   depends_on = [aws_internet_gateway.main]
# }

# resource "aws_route_table" "private" {
#   vpc_id = aws_vpc.main.id

#   route {
#     cidr_block     = "0.0.0.0/0"
#     nat_gateway_id = aws_nat_gateway.main.id
#   }

#   tags = {
#     Name = "private_route_table"
#   }
# }

# resource "aws_route_table_association" "private_subnet_other" {
#   subnet_id      = aws_subnet.other.id
#   route_table_id = aws_route_table.private.id
# }

# resource "aws_route_table_association" "private_subnet_other_2" {
#   subnet_id      = aws_subnet.other_2.id
#   route_table_id = aws_route_table.private.id
# }