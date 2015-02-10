package raws

import (
	"fmt"
	"testing"

	codaws "github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/gen/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccVpc_basic(t *testing.T) {
	var vpc ec2.VPC

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccVpcConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("aws_vpc.foo", &vpc),
					testAccCheckVpcCidr(&vpc, "10.1.0.0/16"),
					resource.TestCheckResourceAttr(
						"aws_vpc.foo", "cidr_block", "10.1.0.0/16"),
				),
			},
		},
	})
}

func TestAccVpc_dedicatedTenancy(t *testing.T) {
	var vpc ec2.VPC

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccVpcDedicatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("aws_vpc.bar", &vpc),
					resource.TestCheckResourceAttr(
						"aws_vpc.bar", "instance_tenancy", "dedicated"),
				),
			},
		},
	})
}

func TestAccVpcUpdate(t *testing.T) {
	var vpc ec2.VPC

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccVpcConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("aws_vpc.foo", &vpc),
					testAccCheckVpcCidr(&vpc, "10.1.0.0/16"),
					resource.TestCheckResourceAttr(
						"aws_vpc.foo", "cidr_block", "10.1.0.0/16"),
				),
			},
			resource.TestStep{
				Config: testAccVpcConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists("aws_vpc.foo", &vpc),
					resource.TestCheckResourceAttr(
						"aws_vpc.foo", "enable_dns_hostnames", "true"),
				),
			},
		},
	})
}

func testAccCheckVpcDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codaConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc" {
			continue
		}
		DescribeVpcOpts := &ec2.DescribeVPCsRequest{
			VPCIDs: []string{rs.Primary.ID},
		}
		// Try to find the VPC
		resp, err := conn.DescribeVPCs(DescribeVpcOpts)
		if err == nil {
			if len(resp.VPCs) > 0 {
				return fmt.Errorf("VPCs still exist.")
			}

			return nil
		}

		// Verify the error is what we want
		ec2err, ok := err.(*codaws.APIError)
		if !ok {
			return err
		}
		if ec2err.Code != "InvalidVpcID.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckVpcCidr(vpc *ec2.VPC, expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		CIDRBlock := vpc.CIDRBlock
		if *CIDRBlock != expected {
			return fmt.Errorf("Bad cidr: %s", vpc.CIDRBlock)
		}

		return nil
	}
}

func testAccCheckVpcExists(n string, vpc *ec2.VPC) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).codaConn
		DescribeVpcOpts := &ec2.DescribeVPCsRequest{
			VPCIDs: []string{rs.Primary.ID},
		}
		resp, err := conn.DescribeVPCs(DescribeVpcOpts)
		if err != nil {
			return err
		}
		if len(resp.VPCs) == 0 {
			return fmt.Errorf("VPC not found")
		}

		*vpc = resp.VPCs[0]

		return nil
	}
}

const testAccVpcConfig = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
}
`

const testAccVpcConfigUpdate = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
	enable_dns_hostnames = true
}
`

const testAccVpcConfigTags = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"

	tags {
		foo = "bar"
	}
}
`

const testAccVpcConfigTagsUpdate = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"

	tags {
		bar = "baz"
	}
}
`
const testAccVpcDedicatedConfig = `
resource "aws_vpc" "bar" {
	instance_tenancy = "dedicated"

	cidr_block = "10.2.0.0/16"
}
`
