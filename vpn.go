package window

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type (
	VPNConnection struct {
		// The configuration information for the VPN connection's customer gateway (in
		// the native XML format). This element is always present in the CreateVpnConnection
		// response; however, it's present in the DescribeVpnConnections response only
		// if the VPN connection is in the pending or available state.
		CustomerGatewayConfiguration string

		// The ID of the customer gateway at your end of the VPN connection.
		CustomerGatewayId string

		// The VPN connection options.
		Options *ec2.VpnConnectionOptions

		// The static routes associated with the VPN connection.
		Routes []*ec2.VpnStaticRoute

		// The current state of the VPN connection.
		State string

		// Any tags assigned to the VPN connection.
		Tags []*ec2.Tag

		// The type of VPN connection.
		Type string

		// Information about the VPN tunnel.
		VgwTelemetry []*ec2.VgwTelemetry

		// The ID of the VPN connection.
		VpnConnectionId string

		// The ID of the virtual private gateway at the AWS side of the VPN connection.
		VpnGatewayId string

		Name                       string
		Id                         string
		VPNConnectionConfiguration *VPNConnectionConfiguration
		VPGateway                  *VPGateway
		CustomerGateway            *CustomerGateway
		CloudWatchAlarms           []*CloudWatchAlarm
	}

	VPNConnectionByNameAsc []*VPNConnection
)

func (a VPNConnectionByNameAsc) Len() int      { return len(a) }
func (a VPNConnectionByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a VPNConnectionByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func (vpn *VPNConnection) String() string {
	if vpn != nil {
		return vpn.Name
	}
	return ""
}

func LoadVPNConnections(input *ec2.DescribeVpnConnectionsInput) (map[string]*VPNConnection, error) {

	resp, err := EC2Client.DescribeVpnConnections(input)
	if err != nil {
		return nil, err
	}

	vpns := make(map[string]*VPNConnection, len(resp.VpnConnections))

	for _, ec2vpn := range resp.VpnConnections {
		vpn := &VPNConnection{
			CustomerGatewayConfiguration: aws.StringValue(ec2vpn.CustomerGatewayConfiguration),
			CustomerGatewayId:            aws.StringValue(ec2vpn.CustomerGatewayId),
			Options:                      ec2vpn.Options,
			Routes:                       ec2vpn.Routes,
			State:                        aws.StringValue(ec2vpn.State),
			Tags:                         ec2vpn.Tags,
			Type:                         aws.StringValue(ec2vpn.Type),
			VgwTelemetry:                 ec2vpn.VgwTelemetry,
			VpnConnectionId:              aws.StringValue(ec2vpn.VpnConnectionId),
			VpnGatewayId:                 aws.StringValue(ec2vpn.VpnGatewayId),
		}
		vpn.Name = TagOrDefault(vpn.Tags, "Name", vpn.VpnConnectionId)
		vpn.Id = "vpn:" + vpn.VpnConnectionId
		vpn.VPNConnectionConfiguration, err = ParseVPNConnectionConfiguration(vpn.CustomerGatewayConfiguration)
		if err != nil {
			log.Println(err)
		}
		vpns[vpn.VpnConnectionId] = vpn
	}

	return vpns, nil

}

func (vpn *VPNConnection) Inactive() bool {
	return vpn.CustomerGateway == nil ||
		vpn.VPGateway == nil ||
		vpn.CustomerGateway.Inactive() ||
		vpn.VPGateway.Inactive()
}
