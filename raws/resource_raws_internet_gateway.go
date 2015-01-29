package raws

import (
	"log"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	codaws "github.com/stripe/aws-go/aws"
	"github.com/stripe/aws-go/gen/ec2"
)

func resourceRawsInternetGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceRawsInternetGatewayCreate,
		Read:   resourceRawsInternetGatewayRead,
		Update: resourceRawsInternetGatewayUpdate,
		Delete: resourceRawsInternetGatewayDelete,

		Schema: map[string]*schema.Schema{
			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRawsInternetGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceRawsInternetGatewayRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceRawsInternetGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceRawsInternetGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
