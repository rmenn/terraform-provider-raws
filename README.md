# terraform-provider-raws
###terraform-provider-raws
Wrote this up since i need to bring up infrastructure in eu-central-1 which goamz does not support due to V4 Signing. Mitchell Hashimoto mentioned that the are planning to move to aws-go, so this is a POC so i can get my Infra up and running.

###Note
Highly Untested.
 
Uses [aws-go], currently supports 
* VPC
* Subnets
* Route Tables ( WIP )
* Route Table Association
* Security Group ( Pending )
* Internet Gateway ( Pending )

Supposed to work with all operations that is supported by TF on VPC and Subnets
```
provider "raws" {
    access_key = "XXXX"
    secret_key = "YYYY"
    region = "eu-central-1"
}

resource "raws_vpc" "main" {
    cidr_block = "10.0.0.0/16"
    instance_tenancy = "default"
}

resource "raws_subnet" "test" {
    vpc_id = "${raws_vpc.main.id}"
    cidr_block = "10.0.1.0/24"
    availability_zone = "eu-central-1a"
}
```
[aws-go]: https://github.com/stripe/aws-go
