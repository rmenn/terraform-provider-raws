package raws

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	codaws "github.com/stripe/aws-go/aws"
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

func resourceRawsRouteTableAssociationRead(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).codaConn
	rtRaw, _, err := resourceAwsRouteTableStateRefreshFunc(ec2conn, d.Get("route_table_id").(string))()
	if err != nil {
		return err
	}
	if rtRaw == nil {
		return nil
	}
	rt := rtRaw.(*ec2.RouteTable)
	found := false
	routeId := d.Id()
	for _, a := range rt.Associations {
		if a.RouteTableAssociationID == &routeId {
			found = true
			d.Set("subnet_id", a.SubnetID)
			break
		}
	}
	if !found {
		d.SetId("")
	}
	return nil
}

func resourceRawsRouteTableAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).codaConn
	subnetId := d.Get("subnet_id").(string)
	routeTableId := d.Get("route_table_id").(string)
	AssociationId := d.Id()
	log.Printf("[INFO] Creating route table association: %s => %s", subnetId, routeTableId)
	ReplaceRouteOpts := &ec2.ReplaceRouteTableAssociationRequest{
		AssociationID: &AssociationId,
		RouteTableID:  &routeTableId,
	}
	resp, err := ec2conn.ReplaceRouteTableAssociation(ReplaceRouteOpts)
	if err != nil {
		ec2err, ok := err.(*codaws.APIError)
		if ok && ec2err.Code == "InvalidAssociationID.NotFound" {
			return resourceRawsRouteTableAssociationCreate(d, meta)
		}
		return err
	}

	d.SetId(*resp.NewAssociationID)
	log.Printf("[INFO] Association ID: %s", d.Id())

	return nil
}

func resourceRawsRouteTableAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).codaConn
	AccocId := d.Id()
	log.Printf("[INFO] Deleting route table association: %s", d.Id())
	DisaccocRouteTableOpts := &ec2.DisassociateRouteTableRequest{
		AssociationID: &AccocId,
	}
	if err := ec2conn.DisassociateRouteTable(DisaccocRouteTableOpts); err != nil {
		ec2err, ok := err.(*codaws.APIError)
		if ok && ec2err.Code == "InvalidAssociationID.NotFound" {
			return nil
		}

		return fmt.Errorf("Error deleting Route Association: %s", err)
	}
	return nil
}
