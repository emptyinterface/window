package window

import (
	"net"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func InTags(tags []*ec2.Tag, value string) bool {
	for _, tag := range tags {
		if tag.Value != nil && strings.Contains(*tag.Value, value) {
			return true
		}
	}
	return false
}

func TagOrDefault(tags []*ec2.Tag, key string, defaults ...string) string {
	for _, tag := range tags {
		if tag != nil && tag.Key != nil && tag.Value != nil && strings.EqualFold(*tag.Key, key) {
			return *tag.Value
		}
	}
	for _, def := range defaults {
		if len(def) > 0 {
			return def
		}
	}
	return ""
}

func TagDescriptionOrDefault(tags []*autoscaling.TagDescription, key string, defaults ...string) string {
	for _, tag := range tags {
		if tag.Key != nil && tag.Value != nil && strings.EqualFold(*tag.Key, key) {
			return *tag.Value
		}
	}
	for _, def := range defaults {
		if len(def) > 0 {
			return def
		}
	}
	return ""
}

var sizeNames []string = []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}

func humanBytes(b uint64, precision int) string {
	const divisor = 1024
	var i int
	var n = float64(b)
	for ; n >= divisor; i, n = i+1, n/divisor {
	}
	return strconv.FormatFloat(n, 'f', precision, 64) + sizeNames[i]
}

func roundTime(d time.Duration) time.Duration {
	switch {
	case d > time.Second:
		d = (d / time.Second) * time.Second
	case d > time.Millisecond:
		d = (d / time.Millisecond) * time.Millisecond
	case d > time.Microsecond:
		d = (d / time.Microsecond) * time.Microsecond
	}
	return d
}

func ElasticCacheClusterInSlice(haystack []*ElasticCacheCluster, needle *ElasticCacheCluster) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
func ACLInSlice(haystack []*ACL, needle *ACL) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
func AMIInSlice(haystack []*AMI, needle *AMI) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
func AutoScalingGroupInSlice(haystack []*AutoScalingGroup, needle *AutoScalingGroup) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
func AvailabilityZoneInSlice(haystack []*AvailabilityZone, needle *AvailabilityZone) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
func CustomerGatewayInSlice(haystack []*CustomerGateway, needle *CustomerGateway) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
func ELBInSlice(haystack []*ELB, needle *ELB) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
func ENIInSlice(haystack []*ENI, needle *ENI) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
func InternetGatewayInSlice(haystack []*InternetGateway, needle *InternetGateway) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
func InstanceInSlice(haystack []*Instance, needle *Instance) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
func DBInstanceInSlice(haystack []*DBInstance, needle *DBInstance) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
func RouteTableInSlice(haystack []*RouteTable, needle *RouteTable) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
func SecurityGroupInSlice(haystack []*SecurityGroup, needle *SecurityGroup) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
func SubnetInSlice(haystack []*Subnet, needle *Subnet) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
func VPCInSlice(haystack []*VPC, needle *VPC) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
func VPCEndpointInSlice(haystack []*VPCEndpoint, needle *VPCEndpoint) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
func VPCPeeringConnectionInSlice(haystack []*VPCPeeringConnection, needle *VPCPeeringConnection) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
func VPGatewayInSlice(haystack []*VPGateway, needle *VPGateway) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}
func VPNConnectionInSlice(haystack []*VPNConnection, needle *VPNConnection) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}

func string_less_than(a, b string) bool {

	for _, ar := range a {
		br, n := utf8.DecodeRuneInString(b)
		if n == 0 {
			return false
		}
		b = b[n:]
		if ar == br {
			continue
		}
		if unicode.IsLower(ar) {
			if unicode.IsLower(br) {
				return ar < br
			}
			br = unicode.ToLower(br)
		} else if unicode.IsLower(br) {
			br = unicode.ToUpper(br)
		}
		if ar == br {
			continue
		}
		return ar < br
	}
	return true

}

func cidr_less_than(anet, bnet *net.IPNet) bool {

	a, b := anet.IP, bnet.IP

	var n int
	if len(a) > len(b) {
		n = len(a)
	} else {
		n = len(b)
	}

	for i := 0; i < n; i++ {
		if a[i] == b[i] {
			continue
		}
		return a[i] < b[i]
	}

	return true

}
