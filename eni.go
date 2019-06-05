package window

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type (
	ENI struct {
		// The association information for an Elastic IP associated with the network
		// interface.
		Association *ec2.NetworkInterfaceAssociation

		// The network interface attachment.
		Attachment *ec2.NetworkInterfaceAttachment

		// The Availability Zone.
		AvailabilityZone string

		// A description.
		Description string

		// Any security groups for the network interface.
		Groups []*ec2.GroupIdentifier

		// The MAC address.
		MacAddress string

		// The ID of the network interface.
		NetworkInterfaceId string

		// The AWS account ID of the owner of the network interface.
		OwnerId string

		// The private DNS name.
		PrivateDnsName string

		// The IP address of the network interface within the subnet.
		PrivateIpAddress string

		// The private IP addresses associated with the network interface.
		PrivateIpAddresses []*ec2.NetworkInterfacePrivateIpAddress

		// The ID of the entity that launched the instance on your behalf (for example,
		// AWS Management Console or Auto Scaling).
		RequesterId string

		// Indicates whether the network interface is being managed by AWS.
		RequesterManaged bool

		// Indicates whether traffic to or from the instance is validated.
		SourceDestCheck bool

		// The status of the network interface.
		Status string

		// The ID of the subnet.
		SubnetId string

		// Any tags assigned to the network interface.
		TagSet []*ec2.Tag

		// The ID of the VPC.
		VpcId string

		Name           string
		Id             string
		State          string
		VPC            *VPC
		SecurityGroups []*SecurityGroup
	}

	ENIByNameAsc []*ENI
)

func (a ENIByNameAsc) Len() int      { return len(a) }
func (a ENIByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ENIByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func LoadENIs(input *ec2.DescribeNetworkInterfacesInput) (map[string]*ENI, error) {

	resp, err := EC2Client.DescribeNetworkInterfaces(input)
	if err != nil {
		return nil, err
	}

	enis := make(map[string]*ENI, len(resp.NetworkInterfaces))

	for _, ec2eni := range resp.NetworkInterfaces {
		eni := &ENI{
			Association:        ec2eni.Association,
			Attachment:         ec2eni.Attachment,
			AvailabilityZone:   aws.StringValue(ec2eni.AvailabilityZone),
			Description:        aws.StringValue(ec2eni.Description),
			Groups:             ec2eni.Groups,
			MacAddress:         aws.StringValue(ec2eni.MacAddress),
			NetworkInterfaceId: aws.StringValue(ec2eni.NetworkInterfaceId),
			OwnerId:            aws.StringValue(ec2eni.OwnerId),
			PrivateDnsName:     aws.StringValue(ec2eni.PrivateDnsName),
			PrivateIpAddress:   aws.StringValue(ec2eni.PrivateIpAddress),
			PrivateIpAddresses: ec2eni.PrivateIpAddresses,
			RequesterId:        aws.StringValue(ec2eni.RequesterId),
			RequesterManaged:   aws.BoolValue(ec2eni.RequesterManaged),
			SourceDestCheck:    aws.BoolValue(ec2eni.SourceDestCheck),
			Status:             aws.StringValue(ec2eni.Status),
			SubnetId:           aws.StringValue(ec2eni.SubnetId),
			TagSet:             ec2eni.TagSet,
			VpcId:              aws.StringValue(ec2eni.VpcId),
		}
		eni.Name = TagOrDefault(eni.TagSet, "Name", eni.NetworkInterfaceId)
		eni.Id = "eni:" + eni.NetworkInterfaceId
		eni.State = eni.Status
		enis[eni.NetworkInterfaceId] = eni
	}

	return enis, nil

}

func (eni *ENI) Inactive() bool {
	return eni.Status == "detached"
}
