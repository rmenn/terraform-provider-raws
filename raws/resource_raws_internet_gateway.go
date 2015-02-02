package raws

import (
	"fmt"
	"log"
	"time"

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
	ec2conn := meta.(*AWSClient).codaConn
	igRaw, _, err := IGStateRefreshFunc(ec2conn, d.Id())()
	if err != nil {
		return err
	}
	if igRaw == nil {
		d.SetId("")
		return nil
	}
	ig := igRaw.(*ec2.InternetGateway)
	d.Set("vpc_id", *ig.Attachments[0].VPCID)
	return nil
}

func resourceRawsInternetGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	if err := resourceAwsInternetGatewayDetach(d, meta); err != nil {
		return err
	}
	return resourceAwsInternetGatewayAttach(d, meta)
}

func resourceRawsInternetGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).codaConn
	if err := resourceAwsInternetGatewayDetach(d, meta); err != nil {
		return err
	}
	log.Printf("[INFO] Deleting Internet Gateway: %s", d.Id())
	return resource.Retry(5*time.Minute, func() error {
		IgId := d.Id()
		DelIGOpts := &ec2.DeleteInternetGatewayRequest{
			InternetGatewayID: &IgId,
		}
		err := ec2conn.DeleteInternetGateway(DelIGOpts)
		if err != nil {
			ec2err, ok := err.(*codaws.APIError)
			if !ok {
				return err
			}
			switch ec2err.Code {
			case "InvalidInternetGatewayID.NotFound":
				return nil
			case "DependencyViolation":
				return err
			default:
				return resource.RetryError{err}
			}
		}
		return fmt.Errorf("Error deleting internet gateway: %s", err)
	})
	return nil
}

func resourceAwsInternetGatewayAttach(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).codaConn
	IgId := d.Id()
	VpcId := d.Get("vpc_id").(string)
	AttachIGOpts := &ec2.AttachInternetGatewayRequest{
		InternetGatewayID: &IgId,
		VPCID:             &VpcId,
	}
	if err := ec2conn.AttachInternetGateway(AttachIGOpts); err != nil {
		return err
	}
	log.Printf("[DEBUG] Waiting for internet gateway (%s) to Attach", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending: []string{"detached", "attaching"},
		Target:  "available",
		Refresh: IGAttachStateRefreshFunc(ec2conn, d.Id(), "available"),
		Timeout: 1 * time.Minute,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for internet gateway (%s) to Attach: %s", d.Id(), err)
	}
	return nil
}

func resourceAwsInternetGatewayDetach(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).codaConn
	wait := true
	IgId := d.Id()
	VpcId := d.Get("vpc_id").(string)
	DetachIGOpts := &ec2.DetachInternetGatewayRequest{
		InternetGatewayID: &IgId,
		VPCID:             &VpcId,
	}
	log.Printf("[INFO] Detaching Internet Gateway '%s' from VPC '%s'", d.Id(), d.Get("vpc_id").(string))
	if err := ec2conn.DetachInternetGateway(DetachIGOpts); err != nil {
		ec2err, ok := err.(*codaws.APIError)
		if ok {
			if ec2err.Code == "InvalidInternetGatewayID.NotFound" {
				err = nil
				wait = false
			} else if ec2err.Code == "Gateway.NotAttached" {
				err = nil
				wait = false
			}
		}
		if err != nil {
			return err
		}
	}
	if !wait {
		return nil
	}
	log.Printf("[DEBUG] Waiting for internet gateway (%s) to detach", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending: []string{"attached", "detaching", "available"},
		Target:  "detached",
		Refresh: IGAttachStateRefreshFunc(ec2conn, d.Id(), "detached"),
		Timeout: 1 * time.Minute,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("Error waiting for internet gateway (%s) to detach: %s", d.Id(), err)
	}
	return nil
}

func IGStateRefreshFunc(ec2conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		DescribeIGWOpts := &ec2.DescribeInternetGatewaysRequest{
			InternetGatewayIDs: []string{id},
		}
		resp, err := ec2conn.DescribeInternetGateways(DescribeIGWOpts)
		if err != nil {
			ec2err, ok := err.(*codaws.APIError)
			if ok && ec2err.Code == "InvalidInternetGatewayID.NotFound" {
				resp = nil
			} else {
				log.Printf("[ERROR] Error on IGStateRefresh: %s", err)
				return nil, "", err
			}
		}
		if resp == nil {
			return nil, "", nil
		}
		ig := &resp.InternetGateways[0]
		return ig, "available", nil
	}
}

func IGAttachStateRefreshFunc(conn *ec2.EC2, id string, expected string) resource.StateRefreshFunc {
	var start time.Time
	return func() (interface{}, string, error) {
		if start.IsZero() {
			start = time.Now()
		}
		DescribeIGWOpts := &ec2.DescribeInternetGatewaysRequest{
			InternetGatewayIDs: []string{id},
		}
		resp, err := conn.DescribeInternetGateways(DescribeIGWOpts)
		if err != nil {
			ec2err, ok := err.(*codaws.APIError)
			if ok && ec2err.Code == "InvalidInternetGatewayID.NotFound" {
				resp = nil
			} else {
				log.Printf("[ERROR] Error on IGStateRefresh: %s", err)
				return nil, "", err
			}
		}
		if resp == nil {
			return nil, "", nil
		}
		ig := &resp.InternetGateways[0]
		if time.Now().Sub(start) > 10*time.Second {
			return ig, expected, nil
		}
		if len(ig.Attachments) == 0 {
			return ig, "detached", nil
		}
		return ig, *ig.Attachments[0].State, nil
	}
	return nil
}
