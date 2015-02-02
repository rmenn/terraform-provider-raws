package raws

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	codaws "github.com/stripe/aws-go/aws"
	"github.com/stripe/aws-go/gen/ec2"
)

func resourceRawsRouteTable() *schema.Resource {
	return &schema.Resource{
		Create: resourceRawsRouteTableCreate,
		Read:   resourceRawsRouteTableRead,
		Update: resourceRawsRouteTableUpdate,
		Delete: resourceRawsRouteTableDelete,

		Schema: map[string]*schema.Schema{
			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"route": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr_block": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"gateway_id": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},

						"instance_id": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
				Set: resourceAwsRouteTableHash,
			},
		},
	}
}

func resourceRawsRouteTableCreate(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).codaConn
	vpcId := d.Get("vpc_id").(string)
	CreateRouteOpts := &ec2.CreateRouteTableRequest{
		VPCID: &vpcId,
	}
	log.Printf("[DEBUG] RouteTable create config: %#v", CreateRouteOpts)
	resp, err := ec2conn.CreateRouteTable(CreateRouteOpts)
	if err != nil {
		return fmt.Errorf("Error creating route table: %s", err)
	}
	log.Printf("[DEBUG] Waiting for route table (%s) to become available", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  "ready",
		Refresh: resourceAwsRouteTableStateRefreshFunc(ec2conn, d.Id()),
		Timeout: 1 * time.Minute,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for route table (%s) to become available: %s", d.Id(), err)
	}
	return resourceRawsRouteTableUpdate(d, meta)
}

func resourceRawsRouteTableRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRawsRouteTableUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRawsRouteTableDelete(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).codaConn
	rtRaw, _, err := resourceAwsRouteTableStateRefreshFunc(ec2conn, d.Id())()
	if err != nil {
		return err
	}
	if rtRaw == nil {
		return nil
	}
	rt := rtRaw.(*ec2.RouteTable)

	for _, a := range rt.Associations {
		log.Printf("[INFO] Disassociating association: %s", a.RouteTableAssociationID)
		DisaccocRouteTableOpts := &ec2.DisassociateRouteTableRequest{
			AssociationID: a.RouteTableAssociationID,
		}
		if err := ec2conn.DisassociateRouteTable(DisaccocRouteTableOpts); err != nil {
			ec2err, ok := err.(*codaws.APIError)
			if ok && ec2err.Code == "InvalidAssociationID.NotFound" {
				return nil
			}
		}
	}
	routeId := d.Id()
	log.Printf("[INFO] Deleting Route Table: %s", d.Id())
	DelRouteOpts := &ec2.DeleteRouteTableRequest{
		RouteTableID: &routeId,
	}
	if err := ec2conn.DeleteRouteTable(DelRouteOpts); err != nil {
		ec2err, ok := err.(*codaws.APIError)
		if ok && ec2err.Code == "InvalidRouteTableID.NotFound" {
			return nil
		}
		return fmt.Errorf("Error deleting route table: %s", err)
	}

	log.Printf("[DEBUG] Waiting for route table (%s) to become destroyed", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending: []string{"ready"},
		Target:  "",
		Refresh: resourceAwsRouteTableStateRefreshFunc(ec2conn, d.Id()),
		Timeout: 1 * time.Minute,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for route table (%s) to become destroyed: %s", d.Id(), err)
	}
	return nil
}

func resourceAwsRouteTableHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["cidr_block"].(string)))

	if v, ok := m["gateway_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["instance_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	return hashcode.String(buf.String())
}

func resourceAwsRouteTableStateRefreshFunc(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		DescribeRouteOpts := &ec2.DescribeRouteTablesRequest{
			RouteTableIDs: []string{id},
		}
		resp, err := conn.DescribeRouteTables(DescribeRouteOpts)
		if err != nil {
			if ec2err, ok := err.(*codaws.APIError); ok && ec2err.Code == "InvalidRouteTableID.NotFound" {
				resp = nil
			} else {
				log.Printf("Error on Route Table State Refresh: %s", err)
				return nil, "", err
			}
		}

		if resp == nil {
			return nil, "", nil
		}

		route := &resp.RouteTables[0]
		return route, "ready", nil
	}
}
