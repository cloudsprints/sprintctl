**Ticket Type:** Task  
**Title:** For Each  
**Project:** Cloud Network Infrastructure Deployment  
**Assignee:** You  
**Reporter:** Derek Morgan  
**Priority:** High  
**Labels:** Terraform, AWS, Locals  
**Epic Link:** AWS VPC Expansion  
**Sprint:** Sprint 01/for\_each

**Lab Setup**
This lab uses Localstack to simulate an AWS environment. Localstack is already preinstalled in your environment. You don't need keys or to configure the provider. If you'd like to use your own account, feel free to specify your provider configuration and run `unset aws` and `unset terraform` to decouple them from Localstack. 

**Description:**

The cloud infrastructure team needs to deploy multiple VPCs dynamically. Each VPC also needs a subnet attached to it. All attributes for these resources must be defined within a `locals.tf` file. The `main.tf` file should only have a single resource defined for the VPCs and one for the subnets. If you were to hardcode everything, it would look like this: 

```
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

```

**Acceptance Criteria:**

> **Note:** If the `terraform validate` command fails, all tasks in the lab will fail!

1\. Create a `locals.tf` file with the necessary `locals` data structure to store the needed information.   
2\. Create the AWS VPC resource and reference the items in the data structure as needed.   
3\. Create the AWS subnet and reference the items in the data structure as needed.   
4\. Use `for_each` to create all necessary resources defined in `locals.tf` while only defining each resource once in `main.tf`  
5\. The `locals` block should have a `vpcs` map that contains the VPC information.   
6\. Within the `vpcs` map, there should be the VPC names, `infra-dev` and `infra-prod`  
7\. Each VPC name should have a nested map for the rest of the information. You will need the `description`, `cidr_block`, `instance_tenancy`, `enable_dns_support`, `subnet_cidr`, and `availability_zone` information in each map. 

**Implementation Notes:**

Feel free to use code from previous labs. The values aren't as important as the concepts. Remember, don't use a `name` attribute in the locals block. Instead, each VPC should have the name of the VPC set as the key that represents that collection of VPC information. 

- <a href="https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/vpc" target="_blank">AWS VPC Resource Documentation</a>  
- <a href="https://developer.hashicorp.com/terraform/language/values/locals" target="_blank">Local Values Guide</a>  
- <a href="https://developer.hashicorp.com/terraform/language/meta-arguments/for_each" target="_blank">for_each Meta-Argument</a>
