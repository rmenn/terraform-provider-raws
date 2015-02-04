package raws

import (
	"github.com/AdRoll/goamz/ec2"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceRawsSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceRawsSecurityGroupCreate,
		Read:   resourceRawsSecurityGroupRead,
		Update: resourceRawsSecurityGroupUpdate,
		Delete: resourceRawsSecurityGroupDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"description": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"ingress": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from_port": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
						},

						"to_port": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
						},

						"protocol": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"cidr_blocks": &schema.Schema{
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},

						"security_groups": &schema.Schema{
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set: func(v interface{}) int {
								return hashcode.String(v.(string))
							},
						},

						"self": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
				Set: resourceAwsSecurityGroupIngressHash,
			},
			"owner_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
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

func resourceAwsSecurityGroupIngressHash(v interface{}) int {
	return nil
}

func SGStateRefreshFunc(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return nil
}
