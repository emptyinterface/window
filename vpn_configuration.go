package window

import (
	"encoding/xml"
	"io"
	"strings"
	"time"
)

type (
	VPNConnectionConfiguration struct {
		XMLName                 xml.Name           `xml:"vpn_connection"`
		Id                      string             `xml:"id,attr"`
		CustomerGatewayId       string             `xml:"customer_gateway_id"`
		VpnGatewayId            string             `xml:"vpn_gateway_id"`
		VpnConnectionType       string             `xml:"vpn_connection_type"`
		VpnConnectionAttributes string             `xml:"vpn_connection_attributes"`
		Tunnels                 []*VPNConfigTunnel `xml:"ipsec_tunnel"`
	}
	VPNConfigTunnel struct {
		CustomerGateway VPNGatewayConfig `xml:"customer_gateway"`
		VPNGateway      VPNGatewayConfig `xml:"vpn_gateway"`
		IKE             struct {
			AuthenticationProtocol string        `xml:"authentication_protocol"`
			EncryptionProtocol     string        `xml:"encryption_protocol"`
			LifetimeSeconds        int           `xml:"lifetime"`
			Lifetime               time.Duration `xml:"-"`
			PerfectForwardSecrecy  string        `xml:"perfect_forward_secrecy"`
			Mode                   string        `xml:"mode"`
			PreSharedKey           string        `xml:"pre_shared_key"`
		} `xml:"ike"`
		IPSec struct {
			Protocol                      string        `xml:"protocol"`
			AuthenticationProtocol        string        `xml:"authentication_protocol"`
			EncryptionProtocol            string        `xml:"encryption_protocol"`
			LifetimeSeconds               int           `xml:"lifetime"`
			Lifetime                      time.Duration `xml:"-"`
			PerfectForwardSecrecy         string        `xml:"perfect_forward_secrecy"`
			Mode                          string        `xml:"mode"`
			ClearDFBit                    bool          `xml:"clear_df_bit"`
			FragmentationBeforeEncryption bool          `xml:"fragmentation_before_encryption"`
			TcpMSSAdjustment              int           `xml:"tcp_mss_adjustment"`
			DeadPeerDetection             struct {
				IntervalSeconds int           `xml:"interval"`
				Interval        time.Duration `xml:"-"`
				Retries         int           `xml:"retries"`
			} `xml:"dead_peer_detection"`
		} `xml:"ipsec"`
	}
	VPNGatewayConfig struct {
		TunnelOutsideAddress struct {
			IPAddress string `xml:"ip_address"`
		} `xml:"tunnel_outside_address"`
		TunnelInsideAddress struct {
			IPAddress   string `xml:"ip_address"`
			NetworkMask string `xml:"network_mask"`
			NetworkCidr string `xml:"network_cidr"`
		} `xml:"tunnel_inside_address"`
	}
)

func ParseVPNConnectionConfiguration(config string) (*VPNConnectionConfiguration, error) {

	vpnconfig := &VPNConnectionConfiguration{}

	if err := xml.NewDecoder(strings.NewReader(config)).Decode(&vpnconfig); err != nil && err != io.EOF {
		return vpnconfig, err
	}

	for _, tunnel := range vpnconfig.Tunnels {
		tunnel.IKE.Lifetime = time.Duration(tunnel.IKE.LifetimeSeconds) * time.Second
		tunnel.IPSec.Lifetime = time.Duration(tunnel.IPSec.LifetimeSeconds) * time.Second
		tunnel.IPSec.DeadPeerDetection.Interval = time.Duration(tunnel.IPSec.DeadPeerDetection.IntervalSeconds) * time.Second
	}

	return vpnconfig, nil

}
