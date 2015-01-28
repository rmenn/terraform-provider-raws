package raws

import (
	"log"

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
