package window

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type (
	NATGateway struct {

		// The date and time the NAT gateway was created.
		CreateTime time.Time

		// The date and time the NAT gateway was deleted, if applicable.
		DeleteTime time.Time

		// If the NAT gateway could not be created, specifies the error code for the
		// failure. (InsufficientFreeAddressesInSubnet | Gateway.NotAttached | InvalidAllocationID.NotFound
		// | Resource.AlreadyAssociated | InternalError | InvalidSubnetID.NotFound)
		FailureCode string

		// If the NAT gateway could not be created, specifies the error message for
		// the failure, that corresponds to the error code.
		//
		//  For InsufficientFreeAddressesInSubnet: "Subnet has insufficient free addresses
		// to create this NAT gateway"
		//
		// For Gateway.NotAttached: "Network vpc-xxxxxxxx has no Internet gateway attached"
		//
		// For InvalidAllocationID.NotFound: "Elastic IP address eipalloc-xxxxxxxx
		// could not be associated with this NAT gateway"
		//
		// For Resource.AlreadyAssociated: "Elastic IP address eipalloc-xxxxxxxx is
		// already associated"
		//
		// For InternalError: "Network interface eni-xxxxxxxx, created and used internally
		// by this NAT gateway is in an invalid state. Please try again."
		//
		// For InvalidSubnetID.NotFound: "The specified subnet subnet-xxxxxxxx does
		// not exist or could not be found."
		FailureMessage string

		// Information about the IP addresses and network interface associated with
		// the NAT gateway.
		NATGatewayAddresses []*ec2.NatGatewayAddress

		// The ID of the NAT gateway.
		NATGatewayId string

		// Reserved. If you need to sustain traffic greater than the documented limits
		// (http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/vpc-nat-gateway.html),
		// contact us through the Support Center (https://console.aws.amazon.com/support/home?).
		ProvisionedBandwidth *ec2.ProvisionedBandwidth

		// The state of the NAT gateway.
		//
		//   pending: The NAT gateway is being created and is not ready to process
		// traffic.
		//
		//   failed: The NAT gateway could not be created. Check the failureCode and
		// failureMessage fields for the reason.
		//
		//   available: The NAT gateway is able to process traffic. This status remains
		// until you delete the NAT gateway, and does not indicate the health of the
		// NAT gateway.
		//
		//   deleting: The NAT gateway is in the process of being terminated and may
		// still be processing traffic.
		//
		//   deleted: The NAT gateway has been terminated and is no longer processing
		// traffic.
		State string

		// The ID of the subnet in which the NAT gateway is located.
		SubnetId string

		// The ID of the VPC in which the NAT gateway is located.
		VpcId string

		Name   string
		Id     string
		Region *Region
		VPC    *VPC
		Subnet *Subnet
	}

	NATGatewaysByNameAsc []*NATGateway
)

func (a NATGatewaysByNameAsc) Len() int      { return len(a) }
func (a NATGatewaysByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a NATGatewaysByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func LoadNATGateways(input *ec2.DescribeNatGatewaysInput) (map[string]*NATGateway, error) {

	resp, err := EC2Client.DescribeNatGateways(input)
	if err != nil {
		return nil, err
	}

	ngs := map[string]*NATGateway{}

	for _, g := range resp.NatGateways {
		ng := &NATGateway{
			CreateTime:           aws.TimeValue(g.CreateTime),
			DeleteTime:           aws.TimeValue(g.DeleteTime),
			FailureCode:          aws.StringValue(g.FailureCode),
			FailureMessage:       aws.StringValue(g.FailureMessage),
			NATGatewayAddresses:  g.NatGatewayAddresses,
			NATGatewayId:         aws.StringValue(g.NatGatewayId),
			ProvisionedBandwidth: g.ProvisionedBandwidth,
			State:                aws.StringValue(g.State),
			SubnetId:             aws.StringValue(g.SubnetId),
			VpcId:                aws.StringValue(g.VpcId),
		}
		ng.Name = ng.NATGatewayId
		ng.Id = "nat:" + ng.NATGatewayId
		ngs[ng.NATGatewayId] = ng
	}

	return ngs, nil

}

func (natgw *NATGateway) Inactive() bool {
	return natgw.State != "available"
}
