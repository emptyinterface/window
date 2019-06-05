package window

import (
	"net"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type (
	Subnet struct {
		// The Availability Zone of the subnet.
		AvailabilityZoneName string

		// The number of unused IP addresses in the subnet. Note that the IP addresses
		// for any stopped instances are considered unavailable.
		AvailableIpAddressCount int64

		// The CIDR block assigned to the subnet.
		CidrBlock string

		// Indicates whether this is the default subnet for the Availability Zone.
		DefaultForAz bool

		// Indicates whether instances launched in this subnet receive a public IP address.
		MapPublicIpOnLaunch bool

		// The current state of the subnet.
		State string

		// The ID of the subnet.
		SubnetId string

		// Any tags assigned to the subnet.
		Tags []*ec2.Tag

		// The ID of the VPC the subnet is in.
		VpcId string

		// new props
		Name             string
		Id               string
		CIDR             *net.IPNet
		VPC              *VPC
		AvailabilityZone *AvailabilityZone
		Instances        []*Instance
		LambdaFunctions  []*LambdaFunction
		RouteTables      []*RouteTable
		ACLs             []*ACL

		InternetGateway       *InternetGateway
		NATInstance           *Instance
		NATGateway            *NATGateway
		VPNConnections        []*VPNConnection
		VPCEndpoints          []*VPCEndpoint
		VPCPeeringConnections []*VPCPeeringConnection
	}

	SubnetByNameAsc []*Subnet
	SubnetByCIDRAsc []*Subnet
)

func (a SubnetByNameAsc) Len() int      { return len(a) }
func (a SubnetByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SubnetByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}
func (a SubnetByCIDRAsc) Len() int      { return len(a) }
func (a SubnetByCIDRAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SubnetByCIDRAsc) Less(i, j int) bool {
	return cidr_less_than(a[i].CIDR, a[j].CIDR)
}

func LoadSubnets(input *ec2.DescribeSubnetsInput) (map[string]*Subnet, error) {

	resp, err := EC2Client.DescribeSubnets(input)
	if err != nil {
		return nil, err
	}

	subnets := make(map[string]*Subnet, len(resp.Subnets))

	for _, ec2subnet := range resp.Subnets {
		subnet := &Subnet{
			AvailabilityZoneName:    aws.StringValue(ec2subnet.AvailabilityZone),
			AvailableIpAddressCount: aws.Int64Value(ec2subnet.AvailableIpAddressCount),
			CidrBlock:               aws.StringValue(ec2subnet.CidrBlock),
			DefaultForAz:            aws.BoolValue(ec2subnet.DefaultForAz),
			MapPublicIpOnLaunch:     aws.BoolValue(ec2subnet.MapPublicIpOnLaunch),
			State:                   aws.StringValue(ec2subnet.State),
			SubnetId:                aws.StringValue(ec2subnet.SubnetId),
			Tags:                    ec2subnet.Tags,
			VpcId:                   aws.StringValue(ec2subnet.VpcId),
		}
		subnet.Name = TagOrDefault(subnet.Tags, "Name", subnet.SubnetId)
		subnet.Id = "subnet:" + subnet.SubnetId
		_, subnet.CIDR, _ = net.ParseCIDR(subnet.CidrBlock)
		subnets[subnet.SubnetId] = subnet
	}

	return subnets, nil

}

func (subnet *Subnet) Inactive() bool {
	return false
}

func (subnet *Subnet) TotalIPs() int {
	bits, size := subnet.CIDR.Mask.Size()
	return 1 << uint(size-bits)
}
