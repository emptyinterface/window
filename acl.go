package window

import (
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type (
	ACL struct {
		// Any associations between the network ACL and one or more acls
		Associations []*ec2.NetworkAclAssociation

		// One or more entries (rules) in the network ACL.
		Entries []*ec2.NetworkAclEntry

		// Indicates whether this is the default network ACL for the VPC.
		IsDefault bool

		// The ID of the network ACL.
		NetworkAclId string

		// Any tags assigned to the network ACL.
		Tags []*ec2.Tag

		// The ID of the VPC for the network ACL.
		VpcId string

		// local props
		Name  string
		Id    string
		State string
	}

	ACLByNameAsc      []*ACL
	EntryByRuleNumber []*ec2.NetworkAclEntry
)

func (a ACLByNameAsc) Len() int      { return len(a) }
func (a ACLByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ACLByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}
func (a EntryByRuleNumber) Len() int      { return len(a) }
func (a EntryByRuleNumber) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a EntryByRuleNumber) Less(i, j int) bool {
	return *a[i].RuleNumber < *a[j].RuleNumber
}

func LoadACLs(input *ec2.DescribeNetworkAclsInput) (map[string]*ACL, error) {

	resp, err := EC2Client.DescribeNetworkAcls(input)
	if err != nil {
		return nil, err
	}

	acls := make(map[string]*ACL, len(resp.NetworkAcls))

	for _, ec2acl := range resp.NetworkAcls {
		acl := &ACL{
			Associations: ec2acl.Associations,
			Entries:      ec2acl.Entries,
			IsDefault:    aws.BoolValue(ec2acl.IsDefault),
			NetworkAclId: aws.StringValue(ec2acl.NetworkAclId),
			Tags:         ec2acl.Tags,
			VpcId:        aws.StringValue(ec2acl.VpcId),
		}
		acl.Name = TagOrDefault(acl.Tags, "Name", acl.NetworkAclId)
		sort.Sort(EntryByRuleNumber(acl.Entries))
		acl.Id = "acl:" + acl.NetworkAclId
		acls[acl.NetworkAclId] = acl
	}

	return acls, nil

}

func (acl *ACL) Inactive() bool {
	return len(acl.Associations) == 0
}
