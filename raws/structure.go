package raws

import (
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/stripe/aws-go/gen/ec2"
)

func expandIPPerms(id string, configured []interface{}) []ec2.IPPermission {
	perms := make([]ec2.IPPermission, len(configured))
	for i, mRaw := range configured {
		var perm ec2.IPPermission
		m := mRaw.(map[string]interface{})
		FromPort := m["from_port"].(int)
		ToPort := m["to_port"].(int)
		Protocol := m["protocol"].(string)
		perm.FromPort = &FromPort
		perm.ToPort = &ToPort
		perm.IPProtocol = &Protocol
		var groups []string
		if raw, ok := m["security_groups"]; ok {
			list := raw.(*schema.Set).List()
			for _, v := range list {
				groups = append(groups, v.(string))
			}
		}
		if v, ok := m["self"]; ok && v.(bool) {
			groups = append(groups, id)
		}
		if len(groups) > 0 {
			perm.UserIDGroupPairs = make([]ec2.UserIDGroupPair, len(groups))
			for i, name := range groups {
				ownerId, id := "", name
				if items := strings.Split(id, "/"); len(items) > 1 {
					ownerId, id = items[0], items[1]
				}

				perm.UserIDGroupPairs[i] = ec2.UserIDGroupPair{
					GroupID: &id,
					UserID:  &ownerId,
				}
			}
		}

		if raw, ok := m["cidr_blocks"]; ok {
			list := raw.([]interface{})
			perm.IPRanges = make([]ec2.IPRange, len(list))
			for i, v := range list {
				Cidr := v.(ec2.IPRange)
				perm.IPRanges[i] = Cidr
			}
		}

		perms[i] = perm
	}

	return perms
}
