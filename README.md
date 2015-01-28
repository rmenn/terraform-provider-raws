# terraform-provider-raws
terraform-provider-raws
Uses [aws-go], currently supports 
* vpc operations ( incomplete ) 
* Subnets
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
