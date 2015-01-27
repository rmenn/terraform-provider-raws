package raws

import (
	"github.com/hashicorp/terraform/helper/schema"
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
			},
		},
	}
}

func resourceRawsRouteTableCreate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRawsRouteTableRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRawsRouteTableUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRawsRouteTableDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
