package raws

import (
	"github.com/hashicorp/terraform/helper/schema"
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

func resourceRawsRouteTableAssociationCreate(d *schema.ResourceData, m interface{}) error {
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
