package window

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type (
	InternetGateway struct {
		Attachments []*ec2.InternetGatewayAttachment

		// The ID of the Internet gateway.
		InternetGatewayId string

		// Any tags assigned to the Internet gateway.
		Tags []*ec2.Tag

		Name  string
		Id    string
		State string
		VPCs  []*VPC
	}

	InternetGatewayByNameAsc []*InternetGateway
)

func (a InternetGatewayByNameAsc) Len() int      { return len(a) }
func (a InternetGatewayByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a InternetGatewayByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func (igw *InternetGateway) String() string {
	if igw != nil {
		return igw.Name
	}
	return ""
}

func LoadInternetGateways(input *ec2.DescribeInternetGatewaysInput) (map[string]*InternetGateway, error) {

	resp, err := EC2Client.DescribeInternetGateways(input)
	if err != nil {
		return nil, err
	}

	igs := make(map[string]*InternetGateway, len(resp.InternetGateways))

	for _, ec2ig := range resp.InternetGateways {
		ig := &InternetGateway{
			Attachments:       ec2ig.Attachments,
			InternetGatewayId: aws.StringValue(ec2ig.InternetGatewayId),
			Tags:              ec2ig.Tags,
		}
		ig.Name = TagOrDefault(ig.Tags, "Name", ig.InternetGatewayId)
		ig.Id = "igw:" + ig.InternetGatewayId
		igs[ig.InternetGatewayId] = ig
	}

	return igs, nil

}

func (ig *InternetGateway) Inactive() bool {
	return len(ig.Attachments) == 0
}
