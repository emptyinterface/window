package window

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type (
	VPCEPolicy struct {
		Statements []struct {
			Action      []string        `json:"-"`
			Resource    []string        `json:"-"`
			ActionRaw   json.RawMessage `json:"Action"`
			ResourceRaw json.RawMessage `json:"Resource"`
			Effect      string          `json:"Effect"`
			Principal   string          `json:"Principal"`
			Sid         string          `json:"Sid"`
		} `json:"Statement"`
		Version string `json:"Version"`
	}

	VPCEndpoint struct {

		// The date and time the VPC endpoint was created.
		CreationTimestamp time.Time

		// The policy document associated with the endpoint.
		PolicyDocument string

		// One or more route tables associated with the endpoint.
		RouteTableIds []string

		// The name of the AWS service to which the endpoint is associated.
		ServiceName string

		// The state of the VPC endpoint.
		State string

		// The ID of the VPC endpoint.
		VpcEndpointId string

		// The ID of the VPC to which the endpoint is associated.
		VpcId string

		Name    string
		Id      string
		Policy  VPCEPolicy
		VPC     *VPC
		Subnets []*Subnet
	}

	VPCEndpointByNameAsc []*VPCEndpoint
)

func (a VPCEndpointByNameAsc) Len() int      { return len(a) }
func (a VPCEndpointByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a VPCEndpointByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func LoadVPCEndpoints(input *ec2.DescribeVpcEndpointsInput) (map[string]*VPCEndpoint, error) {

	resp, err := EC2Client.DescribeVpcEndpoints(input)
	if err != nil {
		return nil, err
	}

	vpces := make(map[string]*VPCEndpoint, len(resp.VpcEndpoints))

	for _, vpce := range resp.VpcEndpoints {
		v := &VPCEndpoint{
			CreationTimestamp: aws.TimeValue(vpce.CreationTimestamp),
			PolicyDocument:    aws.StringValue(vpce.PolicyDocument),
			RouteTableIds:     aws.StringValueSlice(vpce.RouteTableIds),
			ServiceName:       aws.StringValue(vpce.ServiceName),
			State:             aws.StringValue(vpce.State),
			VpcEndpointId:     aws.StringValue(vpce.VpcEndpointId),
			VpcId:             aws.StringValue(vpce.VpcId),
		}
		v.Name = v.VpcEndpointId
		v.Id = "vpce:" + v.VpcEndpointId
		if err := json.NewDecoder(strings.NewReader(v.PolicyDocument)).Decode(&v.Policy); err != nil {
			return nil, err
		}
		for i, _ := range v.Policy.Statements {
			if len(v.Policy.Statements[i].ActionRaw) > 0 {
				if err := json.Unmarshal(v.Policy.Statements[i].ActionRaw, &v.Policy.Statements[i].Action); err != nil {
					var str string
					json.Unmarshal(v.Policy.Statements[i].ActionRaw, &str)
					v.Policy.Statements[i].Action = []string{str}
				}
				v.Policy.Statements[i].ActionRaw = nil
			}
		}
		for i, _ := range v.Policy.Statements {
			if len(v.Policy.Statements[i].ResourceRaw) > 0 {
				if err := json.Unmarshal(v.Policy.Statements[i].ResourceRaw, &v.Policy.Statements[i].Resource); err != nil {
					var str string
					json.Unmarshal(v.Policy.Statements[i].ResourceRaw, &str)
					v.Policy.Statements[i].Resource = []string{str}
				}
				v.Policy.Statements[i].ResourceRaw = nil
			}
		}
		vpces[v.VpcEndpointId] = v
	}

	return vpces, nil

}

func (vpce *VPCEndpoint) Inactive() bool {
	return vpce.State == "deleted" ||
		len(vpce.Subnets) == 0
}
