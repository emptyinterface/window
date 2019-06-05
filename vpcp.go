package window

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type (
	VPCPeeringConnection struct {
		// The information of the peer VPC.
		AccepterVpcInfo *ec2.VpcPeeringConnectionVpcInfo

		// The time that an unaccepted VPC peering connection will expire.
		ExpirationTime time.Time

		// The information of the requester VPC.
		RequesterVpcInfo *ec2.VpcPeeringConnectionVpcInfo

		// The status of the VPC peering connection.
		Status *ec2.VpcPeeringConnectionStateReason

		// Any tags assigned to the resource.
		Tags []*ec2.Tag

		// The ID of the VPC peering connection.
		VpcPeeringConnectionId string

		Name         string
		Id           string
		State        string
		RequesterVPC *VPC
		AccepterVPC  *VPC
		Subnets      []*Subnet
	}

	VPCPeeringConnectionByNameAsc []*VPCPeeringConnection
)

func (a VPCPeeringConnectionByNameAsc) Len() int      { return len(a) }
func (a VPCPeeringConnectionByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a VPCPeeringConnectionByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func LoadVPCPeeringConnections(input *ec2.DescribeVpcPeeringConnectionsInput) (map[string]*VPCPeeringConnection, error) {

	resp, err := EC2Client.DescribeVpcPeeringConnections(input)
	if err != nil {
		return nil, err
	}

	vpcps := make(map[string]*VPCPeeringConnection, len(resp.VpcPeeringConnections))

	for _, vpcp := range resp.VpcPeeringConnections {
		v := &VPCPeeringConnection{
			AccepterVpcInfo:  vpcp.AccepterVpcInfo,
			ExpirationTime:   aws.TimeValue(vpcp.ExpirationTime),
			RequesterVpcInfo: vpcp.RequesterVpcInfo,
			Status:           vpcp.Status,
			Tags:             vpcp.Tags,
			VpcPeeringConnectionId: aws.StringValue(vpcp.VpcPeeringConnectionId),
		}
		v.Name = TagOrDefault(v.Tags, "Name", v.VpcPeeringConnectionId)
		v.Id = "vpcp:" + v.VpcPeeringConnectionId
		v.State = aws.StringValue(v.Status.Code)
		vpcps[v.VpcPeeringConnectionId] = v
	}

	return vpcps, nil

}

func (vpcp *VPCPeeringConnection) Inactive() bool {
	return len(vpcp.Subnets) == 0
}
