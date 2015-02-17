package raws

import (
	"fmt"
	"testing"

	codaws "github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/gen/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSubnet(t *testing.T) {
	var v ec2.Subnet

	testCheck := func(*terraform.State) error {
		CIDRBlock := v.CIDRBlock
		if *CIDRBlock != "10.1.1.0/24" {
			return fmt.Errorf("bad cidr: %s", *CIDRBlock)
		}
		MapPublicIPOnLaunch := v.MapPublicIPOnLaunch
		if *MapPublicIPOnLaunch != true {
			return fmt.Errorf("bad MapPublicIpOnLaunch: %t", *MapPublicIPOnLaunch)
		}

		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckSubnetDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccSubnetConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists(
						"aws_subnet.foo", &v),
					testCheck,
				),
			},
		},
	})
}

func testAccCheckSubnetDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).codaConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_subnet" {
			continue
		}

		// Try to find the resource
		DescribeSubnetOpts := &ec2.DescribeSubnetsRequest{
			SubnetIDs: []string{rs.Primary.ID},
		}
		resp, err := conn.DescribeSubnets(DescribeSubnetOpts)
		if err == nil {
			if len(resp.Subnets) > 0 {
				return fmt.Errorf("still exist.")
			}

			return nil
		}

		// Verify the error is what we want
		ec2err, ok := err.(*codaws.APIError)
		if !ok {
			return err
		}
		if ec2err.Code != "InvalidSubnetID.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckSubnetExists(n string, v *ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).codaConn
		DescribeSubnetOpts := &ec2.DescribeSubnetsRequest{
			SubnetIDs: []string{rs.Primary.ID},
		}
		resp, err := conn.DescribeSubnets(DescribeSubnetOpts)
		if err != nil {
			return err
		}
		if len(resp.Subnets) == 0 {
			return fmt.Errorf("Subnet not found")
		}

		*v = resp.Subnets[0]

		return nil
	}
}

const testAccSubnetConfig = `
resource "aws_vpc" "foo" {
	cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "foo" {
	cidr_block = "10.1.1.0/24"
	vpc_id = "${aws_vpc.foo.id}"
	map_public_ip_on_launch = true
}
`
