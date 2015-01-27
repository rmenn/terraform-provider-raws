package raws

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	codaws "github.com/stripe/aws-go/aws"
	"github.com/stripe/aws-go/gen/ec2"
	"log"
	"time"
)

func resourceRawsSubnet() *schema.Resource {
	return &schema.Resource{
		Create: resourceRawsSubnetCreate,
		Read:   resourceRawsSubnetRead,
		Update: resourceRawsSubnetUpdate,
		Delete: resourceRawsSubnetDelete,

		Schema: map[string]*schema.Schema{
			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},

			"cidr_block": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"availability_zone": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRawsSubnetCreate(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).codaConn
	availability_zone := d.Get("availability_zone").(string)
	cidr_block := d.Get("cidr_block").(string)
	vpc_id := d.Get("vpc_id").(string)
	createOpts := &ec2.CreateSubnetRequest{
		AvailabilityZone: &availability_zone,
		CIDRBlock:        &cidr_block,
		VPCID:            &vpc_id,
	}
	resp, err := ec2conn.CreateSubnet(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating subnet: %s", err)
	}
	subnet := resp.Subnet
	d.SetId(*subnet.SubnetID)
	log.Printf("[INFO] Subnet ID: %s", d.Id())
	log.Printf("[DEBUG] Waiting for subnet (%s) to become available", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  "available",
		Refresh: SubnetStateRefreshFunc(ec2conn, d.Id()),
		Timeout: 10 * time.Minute,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for subnet (%s) to become ready: %s", d.Id(), err)
	}
	return resourceRawsSubnetUpdate(d, meta)
}

func resourceRawsSubnetRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRawsSubnetUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRawsSubnetDelete(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).codaConn
	log.Printf("[INFO] Deleting subnet: %s", d.Id())
	subnet_id := d.Id()
	DelSubnetOpts := &ec2.DeleteSubnetRequest{
		SubnetID: &subnet_id,
	}
	if err := ec2conn.DeleteSubnet(DelSubnetOpts); err != nil {
		ec2err, ok := err.(*codaws.APIError)
		if ok && ec2err.Code == "InvalidSubnetID.NotFound" {
			return nil
		}
		return fmt.Errorf("Error deleting subnet: %s", err)
	}
	return nil
}

func SubnetStateRefreshFunc(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		DescribeSubnetsOpts := &ec2.DescribeSubnetsRequest{
			SubnetIDs: []string{id},
		}
		resp, err := conn.DescribeSubnets(DescribeSubnetsOpts)
		if err != nil {
			if ec2err, ok := err.(*codaws.APIError); ok && ec2err.Code == "InvalidSubnetID.NotFound" {
				resp = nil
			} else {
				log.Printf("Error on SubnettateRefresh: %s", err)
				return nil, "", err
			}
		}
		if resp == nil {
			return nil, "", nil
		}
		subnet := &resp.Subnets[0]
		return subnet, *subnet.State, nil
	}
}
