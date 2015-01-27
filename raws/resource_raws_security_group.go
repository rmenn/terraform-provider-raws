package raws

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceRawsSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceRawsSecurityGroupCreate,
		Read:   resourceRawsSecurityGroupRead,
		Update: resourceRawsSecurityGroupUpdate,
		Delete: resourceRawsSecurityGroupDelete,
	}
}

func resourceRawsSecurityGroupCreate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRawsSecurityGroupRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRawsSecurityGroupUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRawsSecurityGroupDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
