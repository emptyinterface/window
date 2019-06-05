package window

import (
	"net"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type (
	VPC struct {
		// The CIDR block for the VPC.
		CidrBlock string

		// The ID of the set of DHCP options you've associated with the VPC (or default
		// if the default options are associated with the VPC).
		DhcpOptionsId string

		// The allowed tenancy of instances launched into the VPC.
		InstanceTenancy string

		// Indicates whether the VPC is the default VPC.
		IsDefault bool

		// The current state of the VPC.
		// available/pending
		State string

		// Any tags assigned to the VPC.
		Tags []*ec2.Tag

		// The ID of the VPC.
		VpcId string

		Name                  string
		Id                    string
		CIDR                  *net.IPNet
		Region                *Region
		ACLs                  []*ACL
		AMIs                  []*AMI
		AutoScalingGroups     []*AutoScalingGroup
		AvailabilityZones     []*AvailabilityZone
		CustomerGateways      []*CustomerGateway
		DBInstances           []*DBInstance
		ElasticCacheClusters  []*ElasticCacheCluster
		ELBs                  []*ELB
		ENIs                  []*ENI
		Instances             []*Instance
		InternetGateway       *InternetGateway
		NATGateways           []*NATGateway
		LambdaFunctions       []*LambdaFunction
		RouteTables           []*RouteTable
		SecurityGroups        []*SecurityGroup
		Subnets               []*Subnet
		VPCEndpoints          []*VPCEndpoint
		VPCPeeringConnections []*VPCPeeringConnection
		VPGateways            []*VPGateway
		VPNConnections        []*VPNConnection

		azs map[*AvailabilityZone]*AvailabilityZone
	}

	VPCByNameAsc []*VPC
)

func (a VPCByNameAsc) Len() int      { return len(a) }
func (a VPCByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a VPCByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func (vpc *VPC) String() string {
	if vpc != nil {
		return vpc.Name
	}
	return ""
}

func LoadVPCs(input *ec2.DescribeVpcsInput) (map[string]*VPC, error) {

	resp, err := EC2Client.DescribeVpcs(input)
	if err != nil {
		return nil, err
	}

	vpcs := make(map[string]*VPC, len(resp.Vpcs))

	for _, ec2vpc := range resp.Vpcs {
		vpc := &VPC{
			CidrBlock:       aws.StringValue(ec2vpc.CidrBlock),
			DhcpOptionsId:   aws.StringValue(ec2vpc.DhcpOptionsId),
			InstanceTenancy: aws.StringValue(ec2vpc.InstanceTenancy),
			IsDefault:       aws.BoolValue(ec2vpc.IsDefault),
			State:           aws.StringValue(ec2vpc.State),
			Tags:            ec2vpc.Tags,
			VpcId:           aws.StringValue(ec2vpc.VpcId),
			azs:             map[*AvailabilityZone]*AvailabilityZone{},
			Id:              aws.StringValue(ec2vpc.VpcId),
		}
		vpc.Name = TagOrDefault(vpc.Tags, "Name", vpc.VpcId)
		_, vpc.CIDR, _ = net.ParseCIDR(vpc.CidrBlock)
		vpcs[vpc.VpcId] = vpc
	}

	return vpcs, nil

}

func (vpc *VPC) Inactive() bool {
	return false
}

func (vpc *VPC) TotalIPs() int {
	bits, size := vpc.CIDR.Mask.Size()
	return 1 << uint(size-bits)
}
func (vpc *VPC) TotalIPsAllotted() int {
	var allotted int
	for _, subnet := range vpc.Subnets {
		allotted += subnet.TotalIPs()
	}
	return allotted
}
func (vpc *VPC) TotalIPsAvailable() int {
	var available int
	for _, subnet := range vpc.Subnets {
		available += int(subnet.AvailableIpAddressCount)
	}
	return available
}
func (vpc *VPC) TotalIPsInUse() int {
	var allotted, available int
	for _, subnet := range vpc.Subnets {
		allotted += subnet.TotalIPs()
		available += int(subnet.AvailableIpAddressCount)
	}
	return allotted - available
}
