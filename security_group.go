package window

import (
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type (
	SecurityGroup struct {
		// A description of the security group.
		Description string

		// The ID of the security group.
		GroupId string

		// The name of the security group.
		GroupName string

		// One or more inbound rules associated with the security group.
		IpPermissions []*ec2.IpPermission

		// [EC2-VPC] One or more outbound rules associated with the security group.
		IpPermissionsEgress []*ec2.IpPermission

		// The AWS account ID of the owner of the security group.
		OwnerId string

		// Any tags assigned to the security group.
		Tags []*ec2.Tag

		// [EC2-VPC] The ID of the VPC for the security group.
		VpcId string

		Name                 string
		Id                   string
		State                string
		Instances            []*Instance
		ELBs                 []*ELB
		ElasticCacheClusters []*ElasticCacheCluster
		DBInstances          []*DBInstance
		LambdaFunctions      []*LambdaFunction
		Classic              *Classic
		VPCs                 []*VPC
	}

	SecurityGroupSet []*SecurityGroup

	SecurityGroupByNameAsc              []*SecurityGroup
	SecurityGroupIpPermissionsByPortAsc []*ec2.IpPermission
)

func (a SecurityGroupByNameAsc) Len() int      { return len(a) }
func (a SecurityGroupByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SecurityGroupByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func (a SecurityGroupIpPermissionsByPortAsc) Len() int      { return len(a) }
func (a SecurityGroupIpPermissionsByPortAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SecurityGroupIpPermissionsByPortAsc) Less(i, j int) bool {
	if a[i].FromPort != nil && a[j].FromPort != nil {
		return *a[i].FromPort < *a[j].FromPort
	}
	return true
}

func LoadSecurityGroups(input *ec2.DescribeSecurityGroupsInput) (map[string]*SecurityGroup, error) {

	resp, err := EC2Client.DescribeSecurityGroups(input)
	if err != nil {
		return nil, err
	}

	sgs := make(map[string]*SecurityGroup, len(resp.SecurityGroups))

	for _, ec2sg := range resp.SecurityGroups {
		sg := &SecurityGroup{
			Description:         aws.StringValue(ec2sg.Description),
			GroupId:             aws.StringValue(ec2sg.GroupId),
			GroupName:           aws.StringValue(ec2sg.GroupName),
			IpPermissions:       ec2sg.IpPermissions,
			IpPermissionsEgress: ec2sg.IpPermissionsEgress,
			OwnerId:             aws.StringValue(ec2sg.OwnerId),
			Tags:                ec2sg.Tags,
			VpcId:               aws.StringValue(ec2sg.VpcId),
		}
		sg.Name = TagOrDefault(sg.Tags, "Name", sg.GroupName, sg.GroupId)
		sg.Id = "sg:" + sg.GroupId
		sgs[sg.GroupId] = sg
	}

	return sgs, nil

}

func (sg *SecurityGroup) PortsInvolved() []int {

	set := map[int64]struct{}{}

	for _, perm := range sg.IpPermissions {
		if perm.FromPort != nil {
			set[*perm.FromPort] = struct{}{}
		}
		if perm.ToPort != nil {
			set[*perm.ToPort] = struct{}{}
		}
		if perm.FromPort == nil && perm.ToPort == nil && perm.IpProtocol != nil && *perm.IpProtocol == "-1" {
			set[-1] = struct{}{}
		}
	}
	for _, perm := range sg.IpPermissionsEgress {
		if perm.FromPort != nil {
			set[*perm.FromPort] = struct{}{}
		}
		if perm.ToPort != nil {
			set[*perm.ToPort] = struct{}{}
		}
		if perm.FromPort == nil && perm.ToPort == nil && perm.IpProtocol != nil && *perm.IpProtocol == "-1" {
			set[-1] = struct{}{}
		}
	}

	var ports []int
	for i, _ := range set {
		ports = append(ports, int(i))
	}

	sort.Ints(ports)

	return ports

}

func (sg *SecurityGroup) Inactive() bool {
	return len(sg.Instances) == 0 &&
		len(sg.ELBs) == 0 &&
		len(sg.DBInstances) == 0 &&
		len(sg.ElasticCacheClusters) == 0
}

func (set SecurityGroupSet) PortsInvolved() []int {

	var (
		portmap = map[int]struct{}{}
		ports   []int
	)

	for _, sg := range set {
		for _, port := range sg.PortsInvolved() {
			portmap[port] = struct{}{}
		}
	}

	for port, _ := range portmap {
		ports = append(ports, port)
	}

	sort.Ints(ports)

	return ports

}
