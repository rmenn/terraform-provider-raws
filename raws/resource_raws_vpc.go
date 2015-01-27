package raws

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	codaws "github.com/stripe/aws-go/aws"
	"github.com/stripe/aws-go/gen/ec2"
)

func resourceRawsVpc() *schema.Resource {
	return &schema.Resource{
		Create: resourceRawsVpcCreate,
		Read:   resourceRawsVpcRead,
		Update: resourceRawsVpcUpdate,
		Delete: resourceRawsVpcDelete,

		Schema: map[string]*schema.Schema{
			"cidr_block": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"instance_tenancy": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRawsVpcCreate(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).codaConn
	cidr := d.Get("cidr_block").(string)
	instance_tenancy := "default"
	if v := d.Get("instance_tenancy"); v != nil {
		instance_tenancy = v.(string)
	}
	createOpts := &ec2.CreateVPCRequest{
		CIDRBlock:       &cidr,
		InstanceTenancy: &instance_tenancy,
	}
	log.Printf("[DEBUG] VPC create config: %#v", createOpts)
	vpcResp, err := ec2conn.CreateVPC(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating VPC: %s", err)
	}
	vpc := vpcResp.VPC
	log.Printf("[INFO] VPC ID: %s", vpc.VPCID)
	d.SetId(*vpc.VPCID)
	d.Partial(true)
	d.SetPartial("cidr_block")
	// Wait for the VPC to become available
	log.Printf("[DEBUG] Waiting for VPC (%s) to become available", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  "available",
		Refresh: VPCStateRefreshFunc(ec2conn, d.Id()),
		Timeout: 10 * time.Minute,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for VPC (%s) to become available: %s", d.Id(), err)
	}

	// Update our attributes and return
	return resourceRawsVpcUpdate(d, meta)
}

func resourceRawsVpcRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRawsVpcUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRawsVpcDelete(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).codaConn
	vpcID := d.Id()
	DeleteVpcOpts := &ec2.DeleteVPCRequest{
		VPCID: &vpcID,
	}
	log.Printf("[INFO] Deleting VPC: %s", d.Id())
	if err := ec2conn.DeleteVPC(DeleteVpcOpts); err != nil {
		ec2err, ok := err.(*codaws.APIError)
		if ok && ec2err.Code == "InvalidVpcID.NotFound" {
			return nil
		}

		return fmt.Errorf("Error deleting VPC: %s", err)
	}

	return nil
}

// VPCStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch
// a VPC.
func VPCStateRefreshFunc(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		DescribeVpcOpts := &ec2.DescribeVPCsRequest{
			VPCIDs: []string{id},
		}
		resp, err := conn.DescribeVPCs(DescribeVpcOpts)
		if err != nil {
			if ec2err, ok := err.(*codaws.APIError); ok && ec2err.Code == "InvalidVpcID.NotFound" {
				resp = nil
			} else {
				log.Printf("Error on VPCStateRefresh: %s", err)
				return nil, "", err
			}
		}

		if resp == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our instance yet. Return an empty state.
			return nil, "", nil
		}

		vpc := &resp.VPCs[0]
		return vpc, *vpc.State, nil
	}
}
