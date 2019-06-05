package window

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type (
	AvailabilityZone struct {
		// Any messages about the Availability Zone.
		Messages []*ec2.AvailabilityZoneMessage

		// The name of the region.
		RegionName string

		// The state of the Availability Zone.
		State string

		// The name of the Availability Zone.
		ZoneName string

		Name                 string
		Id                   string
		Instances            []*Instance
		ENIs                 []*ENI
		DBInstances          []*DBInstance
		ElasticCacheClusters []*ElasticCacheCluster
		Subnets              []*Subnet
	}

	AvailabilityZoneByNameAsc []*AvailabilityZone
)

func (a AvailabilityZoneByNameAsc) Len() int      { return len(a) }
func (a AvailabilityZoneByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a AvailabilityZoneByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func LoadAvailabilityZones(input *ec2.DescribeAvailabilityZonesInput) (map[string]*AvailabilityZone, error) {

	resp, err := EC2Client.DescribeAvailabilityZones(input)
	if err != nil {
		return nil, err
	}

	azs := make(map[string]*AvailabilityZone, len(resp.AvailabilityZones))

	for _, ec2az := range resp.AvailabilityZones {
		azs[*ec2az.ZoneName] = &AvailabilityZone{
			Messages:   ec2az.Messages,
			RegionName: aws.StringValue(ec2az.RegionName),
			State:      aws.StringValue(ec2az.State),
			ZoneName:   aws.StringValue(ec2az.ZoneName),
			Name:       aws.StringValue(ec2az.ZoneName),
			Id:         aws.StringValue(ec2az.ZoneName),
		}
	}

	return azs, nil

}

func (az *AvailabilityZone) ClassicInstances() []*Instance {
	var insts []*Instance
	for _, inst := range az.Instances {
		if inst.Classic != nil {
			insts = append(insts, inst)
		}
	}
	return insts
}
func (az *AvailabilityZone) VPCInstances(vpc *VPC) []*Instance {
	var insts []*Instance
	for _, inst := range az.Instances {
		if inst.VPC == vpc {
			insts = append(insts, inst)
		}
	}
	return insts
}
