package raws

import (
	"github.com/hashicorp/terraform/helper/schema"
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

func resourceRawsSubnetCreate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRawsSubnetRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRawsSubnetUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRawsSubnetDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
