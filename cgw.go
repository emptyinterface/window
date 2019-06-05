package window

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type (
	CustomerGateway struct {
		// The customer gateway's Border Gateway Protocol (BGP) Autonomous System Number
		// (ASN).
		BgpAsn string

		// The ID of the customer gateway.
		CustomerGatewayId string

		// The Internet-routable IP address of the customer gateway's outside interface.
		IpAddress string

		// The current state of the customer gateway (pending | available | deleting | deleted).
		State string

		// Any tags assigned to the customer gateway.
		Tags []*ec2.Tag

		// The type of VPN connection the customer gateway supports (ipsec.1).
		Type string

		Name string
		Id   string
	}

	CustomerGatewayByNameAsc []*CustomerGateway
)

func (a CustomerGatewayByNameAsc) Len() int      { return len(a) }
func (a CustomerGatewayByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a CustomerGatewayByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func (cgw *CustomerGateway) String() string {
	if cgw != nil {
		return cgw.Name
	}
	return ""
}

func LoadCustomerGateways(input *ec2.DescribeCustomerGatewaysInput) (map[string]*CustomerGateway, error) {

	resp, err := EC2Client.DescribeCustomerGateways(input)
	if err != nil {
		return nil, err
	}

	cgws := make(map[string]*CustomerGateway, len(resp.CustomerGateways))

	for _, ec2cgw := range resp.CustomerGateways {
		cgw := &CustomerGateway{
			BgpAsn:            aws.StringValue(ec2cgw.BgpAsn),
			CustomerGatewayId: aws.StringValue(ec2cgw.CustomerGatewayId),
			IpAddress:         aws.StringValue(ec2cgw.IpAddress),
			State:             aws.StringValue(ec2cgw.State),
			Tags:              ec2cgw.Tags,
			Type:              aws.StringValue(ec2cgw.Type),
			Id:                aws.StringValue(ec2cgw.CustomerGatewayId),
		}
		cgw.Name = TagOrDefault(cgw.Tags, "Name", cgw.CustomerGatewayId)
		cgws[cgw.CustomerGatewayId] = cgw
	}

	return cgws, nil

}

func (cgw *CustomerGateway) Inactive() bool {
	return cgw.State == "deleted"
}
