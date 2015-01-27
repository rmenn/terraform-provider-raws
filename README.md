# terraform-provider-raws
terraform-provider-raws

```
provider "raws" {
    access_key = "XXXX"
    secret_key = "YYYY"
    region = "us-east-1"
}

resource "raws_vpc" "main" {
    cidr_block = "10.0.0.0/16"
    instance_tenancy = "default"
}
```
