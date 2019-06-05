package window

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type (
	RouteTable struct {
		// The associations between the route table and one or more subnets.
		Associations []*ec2.RouteTableAssociation

		// Any virtual private gateway (VGW) propagating routes.
		PropagatingVgws []*ec2.PropagatingVgw

		// The ID of the route table.
		RouteTableId string

		// The routes in the route table.
		Routes []*ec2.Route

		// Any tags assigned to the route table.
		Tags []*ec2.Tag

		// The ID of the VPC.
		VpcId string

		Name  string
		Id    string
		State string
	}

	RouteTableByNameAsc []*RouteTable
)

func (a RouteTableByNameAsc) Len() int      { return len(a) }
func (a RouteTableByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a RouteTableByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func LoadRouteTables(input *ec2.DescribeRouteTablesInput) (map[string]*RouteTable, error) {

	resp, err := EC2Client.DescribeRouteTables(input)
	if err != nil {
		return nil, err
	}

	route_tables := make(map[string]*RouteTable, len(resp.RouteTables))

	for _, ec2rt := range resp.RouteTables {
		route_table := &RouteTable{
			Associations:    ec2rt.Associations,
			PropagatingVgws: ec2rt.PropagatingVgws,
			RouteTableId:    aws.StringValue(ec2rt.RouteTableId),
			Routes:          ec2rt.Routes,
			Tags:            ec2rt.Tags,
			VpcId:           aws.StringValue(ec2rt.VpcId),
		}
		route_table.Name = TagOrDefault(route_table.Tags, "Name", route_table.RouteTableId)
		route_table.Id = "rt:" + route_table.RouteTableId
		route_tables[route_table.RouteTableId] = route_table
	}

	return route_tables, nil

}

func (rt *RouteTable) Inactive() bool {
	return len(rt.Associations) == 0
}
