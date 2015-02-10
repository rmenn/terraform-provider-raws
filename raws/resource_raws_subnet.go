package raws

import (
	"fmt"
	codaws "github.com/awslabs/aws-sdk-go/aws"
	"github.com/awslabs/aws-sdk-go/gen/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
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

			"map_public_ip_on_launch": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
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

func resourceRawsSubnetRead(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).codaConn
	subnetRaw, _, err := SubnetStateRefreshFunc(ec2conn, d.Id())()
	if err != nil {
		return err
	}
	if subnetRaw == nil {
		return nil
	}
	subnet := subnetRaw.(*ec2.Subnet)
	d.Set("availability_zone", subnet.AvailabilityZone)
	d.Set("vpc_id", subnet.VPCID)
	d.Set("cidr_block", subnet.CIDRBlock)
	return nil
}

func resourceRawsSubnetUpdate(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).codaConn
	d.Partial(true)
	modify := false
	subnetId := d.Id()
	if d.HasChange("map_public_ip_on_launch") {
		ModSubnetAttrOpts := &ec2.ModifySubnetAttributeRequest{
			SubnetID: &subnetId,
		}
		if v, ok := d.GetOk("source_dest_check"); ok {
			val := v.(bool)
			ModSubnetAttrOpts.MapPublicIPOnLaunch = &ec2.AttributeBooleanValue{
				Value: &val,
			}
			modify = true
		}
		if modify {
			log.Printf("[INFO] Modifing instance %s: %#v", d.Id(), ModSubnetAttrOpts)
			if err := ec2conn.ModifySubnetAttribute(ModSubnetAttrOpts); err != nil {
				return err
			}
		}
	}
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
