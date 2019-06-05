package window

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type (
	VPGateway struct {
		AvailabilityZone string

		// The current state of the virtual private gateway.
		State string

		// Any tags assigned to the virtual private gateway.
		Tags []*ec2.Tag

		// The type of VPN connection the virtual private gateway supports.
		Type string

		// Any VPCs attached to the virtual private gateway.
		VpcAttachments []*ec2.VpcAttachment

		// The ID of the virtual private gateway.
		VpnGatewayId string

		Name             string
		Id               string
		VPCs             []*VPC
		VPNConnections   []*VPNConnection
		Subnets          []*Subnet
		CloudWatchAlarms []*CloudWatchAlarm
	}

	VPGatewayByNameAsc []*VPGateway
)

func (a VPGatewayByNameAsc) Len() int      { return len(a) }
func (a VPGatewayByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a VPGatewayByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func (vpg *VPGateway) String() string {
	if vpg != nil {
		return vpg.Name
	}
	return ""
}

func LoadVPGateways(input *ec2.DescribeVpnGatewaysInput) (map[string]*VPGateway, error) {

	resp, err := EC2Client.DescribeVpnGateways(input)
	if err != nil {
		return nil, err
	}

	vpgs := make(map[string]*VPGateway, len(resp.VpnGateways))

	for _, ec2vpg := range resp.VpnGateways {
		vpg := &VPGateway{
			AvailabilityZone: aws.StringValue(ec2vpg.AvailabilityZone),
			State:            aws.StringValue(ec2vpg.State),
			Tags:             ec2vpg.Tags,
			Type:             aws.StringValue(ec2vpg.Type),
			VpcAttachments:   ec2vpg.VpcAttachments,
			VpnGatewayId:     aws.StringValue(ec2vpg.VpnGatewayId),
		}
		vpg.Name = TagOrDefault(vpg.Tags, "Name", vpg.VpnGatewayId)
		vpg.Id = "vpg:" + vpg.VpnGatewayId
		vpgs[vpg.VpnGatewayId] = vpg
	}

	return vpgs, nil

}

func (vpg *VPGateway) Inactive() bool {
	return vpg.State == "deleted" ||
		len(vpg.VPCs) == 0 ||
		len(vpg.VPNConnections) == 0
}
