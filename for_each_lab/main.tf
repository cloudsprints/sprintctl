resource "aws_vpc" "infra-dev" {
  cidr_block           = "10.0.0.0/16"
  instance_tenancy     = "default"
  enable_dns_support   = true
  tags = {
    Name        = "infra-dev"
    Description = "Infrastructure Dev VPC"
  }
}

resource "aws_subnet" "infra-dev-subnet" {
  vpc_id            = aws_vpc.infra-dev.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-east-1a"
  tags = {
    Name = "dev-subnet"
  }
}

resource "aws_vpc" "infra-prod" {
  cidr_block           = "10.1.0.0/16"
  instance_tenancy     = "default"
  enable_dns_support   = true
  tags = {
    Name        = "infra-prod"
    Description = "Infrastructure Prod VPC"
  }
}

resource "aws_subnet" "infra-prod-subnet" {
  vpc_id            = aws_vpc.infra-prod.id
  cidr_block        = "10.1.1.0/24"
  availability_zone = "us-east-1a"
  tags = {
    Name = "prod-subnet"
  }
}

