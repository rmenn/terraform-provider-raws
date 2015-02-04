package raws

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	codaws "github.com/stripe/aws-go/aws"
	"github.com/stripe/aws-go/gen/ec2"
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
func resourceRawsSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).codaConn
	SgName := d.Get("name").(string)
	var VpcId string
	var SgDescription string
	if v := d.Get("vpc_id"); v != nil {
		VpcId = v.(string)
	}
	if v := d.Get("description"); v != nil {
		SgDescription = v.(string)
	}
	CreateSgOpts := &ec2.CreateSecurityGroupRequest{
		GroupName:   &SgName,
		VPCID:       &VpcId,
		Description: &SgDescription,
	}
	log.Printf("[DEBUG] Security Group create configuration: %#v", CreateSgOpts)
	resp, err := ec2conn.CreateSecurityGroup(CreateSgOpts)
	if err != nil {
		return fmt.Errorf("Error creating Security Group: %s", err)
	}
	d.SetId(*resp.GroupID)
	log.Printf("[INFO] Security Group ID: %s", d.Id())
	log.Printf("[DEBUG] Waiting for Security Group (%s) to exist", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending: []string{""},
		Target:  "exists",
		Refresh: SGStateRefreshFunc(ec2conn, d.Id()),
		Timeout: 1 * time.Minute,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf(
			"Error waiting for Security Group (%s) to become available: %s",
			d.Id(), err)
	}
	return resourceRawsSecurityGroupUpdate(d, meta)
}

func resourceRawsSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).codaConn
	sgRaw, _, err := SGStateRefreshFunc(ec2conn, d.Id())()
	if err != nil {
		return err
	}
	if sgRaw == nil {
		d.SetId("")
		return nil
	}
	sg := sgRaw.(*ec2.SecurityGroup)
	ingressMap := make(map[string]map[string]interface{})
	for _, perm := range sg.IPPermissions {
		k := fmt.Sprintf("%s-%d-%d", perm.IPProtocol, perm.FromPort, perm.ToPort)
		m, ok := ingressMap[k]
		if !ok {
			m = make(map[string]interface{})
			ingressMap[k] = m
		}
		m["from_port"] = perm.FromPort
		m["to_port"] = perm.ToPort
		m["protocol"] = perm.IPProtocol
		if len(perm.IPRanges) > 0 {
			raw, ok := m["cidr_blocks"]
			if !ok {
				raw = make([]string, 0, len(perm.IPRanges))
			}
			list := raw.([]ec2.IPRange)
			list = append(list, perm.IPRanges...)
			m["cidr_blocks"] = list
		}

		var groups []string
		if len(perm.UserIDGroupPairs) > 0 {
			groups = flattenSecurityGroups(perm.UserIDGroupPairs)
		}
		for i, id := range groups {
			if id == d.Id() {
				groups[i], groups = groups[len(groups)-1], groups[:len(groups)-1]
				m["self"] = true
			}
		}
		if len(groups) > 0 {
			raw, ok := m["security_groups"]
			if !ok {
				raw = make([]string, 0, len(groups))
			}
			list := raw.([]string)

			list = append(list, groups...)
			m["security_groups"] = list
		}
	}
	ingressRules := make([]map[string]interface{}, 0, len(ingressMap))
	for _, m := range ingressMap {
		ingressRules = append(ingressRules, m)
	}

	d.Set("description", sg.Description)
	d.Set("name", sg.GroupName)
	d.Set("vpc_id", sg.VPCID)
	d.Set("owner_id", sg.OwnerID)
	d.Set("ingress", ingressRules)

	return nil
}

func resourceRawsSecurityGroupUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceRawsSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {
	ec2conn := meta.(*AWSClient).codaConn
	log.Printf("[DEBUG] Security Group destroy: %v", d.Id())
	return resource.Retry(5*time.Minute, func() error {
		SgId := d.Id()
		DelSgOpts := &ec2.DeleteSecurityGroupRequest{
			GroupID: &SgId,
		}
		err := ec2conn.DeleteSecurityGroup(DelSgOpts)
		if err != nil {
			ec2err, ok := err.(*codaws.APIError)
			if !ok {
				return err
			}
			switch ec2err.Code {
			case "InvalidGroup.NotFound":
				return nil
			case "DependencyViolation":
				return err
			default:
				return resource.RetryError{err}
			}
		}
		return nil
	})
}

func resourceAwsSecurityGroupIngressHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%d-", m["from_port"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", m["to_port"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", m["protocol"].(string)))
	if v, ok := m["cidr_blocks"]; ok {
		vs := v.([]interface{})
		s := make([]string, len(vs))
		for i, raw := range vs {
			s[i] = raw.(string)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}
	if v, ok := m["security_groups"]; ok {
		vs := v.(*schema.Set).List()
		s := make([]string, len(vs))
		for i, raw := range vs {
			s[i] = raw.(string)
		}
		sort.Strings(s)

		for _, v := range s {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}

	return hashcode.String(buf.String())
}

func SGStateRefreshFunc(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		DescribeSgOpts := &ec2.DescribeSecurityGroupsRequest{
			GroupIDs: []string{id},
		}
		resp, err := conn.DescribeSecurityGroups(DescribeSgOpts)
		if err != nil {
			if ec2err, ok := err.(*codaws.APIError); ok {
				if ec2err.Code == "InvalidSecurityGroupID.NotFound" || ec2err.Code == "InvalidGroup.NotFound" {
					resp = nil
					err = nil
				}
			}
			if err != nil {
				log.Printf("Error on SGStateRefresh: %s", err)
				return nil, "", err
			}
		}
		if resp == nil {
			return nil, "", nil
		}
		group := &resp.SecurityGroups[0]
		return group, "exists", nil
	}
}
