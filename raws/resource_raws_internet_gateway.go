package raws

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
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
	ec2conn := meta.(*AWSClient).codaConn
	log.Printf("[DEBUG] Creating internet gateway")
	resp, err := ec2conn.CreateInternetGateway(nil)
	if err != nil {
		return fmt.Errorf("Error creating internet gateway: %s", err)
	}
	ig := resp.InternetGateway
	d.SetId(*ig.InternetGatewayID)
	log.Printf("[INFO] InternetGateway ID: %s", d.Id())
	return resourceAwsInternetGatewayAttach(d, meta)
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

func resourceAwsInternetGatewayAttach(d *schema.ResourceData, meta interface{}) error {
	return nil
}
