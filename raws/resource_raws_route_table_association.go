package raws

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/stripe/aws-go/gen/ec2"
)

func resourceRawsRouteTableAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceRawsRouteTableAssociationCreate,
		Read:   resourceRawsRouteTableAssociationRead,
		Update: resourceRawsRouteTableAssociationUpdate,
		Delete: resourceRawsRouteTableAssociationDelete,

		Schema: map[string]*schema.Schema{
			"subnet_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"route_table_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceRawsRouteTableAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).codaConn
	subnet_id := d.Get("subnet_id").(string)
	route_table_id := d.Get("route_table_id").(string)
	associateOpts := &ec2.AssociateRouteTableRequest{
		RouteTableID: &route_table_id,
		SubnetID:     &subnet_id,
	}
	resp, err := ec2conn.AssociateRouteTable(associateOpts)
	if err != nil {
		return err
	}
	d.SetId(*resp.AssociationID)
	log.Printf("[INFO] Association ID: %s", d.Id())
	return nil
}

func resourceRawsRouteTableAssociationRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRawsRouteTableAssociationUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRawsRouteTableAssociationDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
