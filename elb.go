package window

import (
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
)

type (
	ELB struct {
		// The Availability Zones for the load balancer.
		AvailabilityZoneNames []string

		// Information about the back-end servers.
		BackendServerDescriptions []*elb.BackendServerDescription

		// The Amazon Route 53 hosted zone associated with the load balancer.
		//
		// For more information, see Using Domain Names With Elastic Load Balancing
		// (http://docs.aws.amazon.com/ElasticLoadBalancing/latest/DeveloperGuide/using-domain-names-with-elb.html)
		// in the Elastic Load Balancing Developer Guide.
		CanonicalHostedZoneName string

		// The ID of the Amazon Route 53 hosted zone name associated with the load balancer.
		CanonicalHostedZoneNameID string

		// The date and time the load balancer was created.
		CreatedTime time.Time

		// The external DNS name of the load balancer.
		DNSName string

		// Information about the health checks conducted on the load balancer.
		HealthCheck *elb.HealthCheck

		// The IDs of the instances for the load balancer.
		ELBInstances []*elb.Instance

		// The listeners for the load balancer.
		ListenerDescriptions []*elb.ListenerDescription

		// The name of the load balancer.
		LoadBalancerName string

		// The policies defined for the load balancer.
		Policies *elb.Policies

		// The type of load balancer. Valid only for load balancers in a VPC.
		//
		// If Scheme is internet-facing, the load balancer has a public DNS name that
		// resolves to a public IP address.
		//
		// If Scheme is internal, the load balancer has a public DNS name that resolves
		// to a private IP address.
		Scheme string

		// The security groups for the load balancer. Valid only for load balancers
		// in a VPC.
		SecurityGroupNames []string

		// The security group that you can use as part of your inbound rules for your
		// load balancer's back-end application instances. To only allow traffic from
		// load balancers, add a security group rule to your back end instance that
		// specifies this source security group as the inbound source.
		SourceSecurityGroupName string

		// The IDs of the subnets for the load balancer.
		SubnetIds []string

		// The ID of the VPC for the load balancer.
		VpcId string // VPCId string

		Name                string
		Id                  string
		State               string
		Region              *Region
		VPC                 *VPC
		Classic             *Classic
		AvailabilityZones   []*AvailabilityZone
		Subnets             []*Subnet
		AutoScalingGroups   []*AutoScalingGroup
		Instances           []*Instance
		SecurityGroups      []*SecurityGroup
		SourceSecurityGroup *SecurityGroup
		CloudWatchAlarms    []*CloudWatchAlarm

		Stats *ELBStats
	}

	ELBPolicies struct {
		LBCookieStickinessPolicies []struct {
			CookieExpirationPeriod int    `json:"CookieExpirationPeriod"`
			PolicyName             string `json:"PolicyName"`
		} `json:"LBCookieStickinessPolicies"`
		OtherPolicies []string `json:"OtherPolicies"`
	}
	ELBRepresentation struct {
		Name string
	}

	ELBByNameAsc []*ELB

	ELBSet []*ELB
)

func (a ELBByNameAsc) Len() int      { return len(a) }
func (a ELBByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ELBByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func LoadELBs(input *elb.DescribeLoadBalancersInput) (map[string]*ELB, error) {

	elbs := map[string]*ELB{}

	if err := ELBClient.DescribeLoadBalancersPages(input, func(page *elb.DescribeLoadBalancersOutput, _ bool) bool {
		for _, awselb := range page.LoadBalancerDescriptions {
			elb := &ELB{
				AvailabilityZoneNames:     aws.StringValueSlice(awselb.AvailabilityZones),
				BackendServerDescriptions: awselb.BackendServerDescriptions,
				CanonicalHostedZoneName:   aws.StringValue(awselb.CanonicalHostedZoneName),
				CanonicalHostedZoneNameID: aws.StringValue(awselb.CanonicalHostedZoneNameID),
				CreatedTime:               aws.TimeValue(awselb.CreatedTime),
				DNSName:                   aws.StringValue(awselb.DNSName),
				HealthCheck:               awselb.HealthCheck,
				ELBInstances:              awselb.Instances,
				ListenerDescriptions:      awselb.ListenerDescriptions,
				LoadBalancerName:          aws.StringValue(awselb.LoadBalancerName),
				Policies:                  awselb.Policies,
				Scheme:                    aws.StringValue(awselb.Scheme),
				SecurityGroupNames:        aws.StringValueSlice(awselb.SecurityGroups),
				SubnetIds:                 aws.StringValueSlice(awselb.Subnets),
				VpcId:                     aws.StringValue(awselb.VPCId),
			}
			if awselb.SourceSecurityGroup != nil {
				elb.SourceSecurityGroupName = aws.StringValue(awselb.SourceSecurityGroup.GroupName)
			}
			elb.Name = elb.LoadBalancerName
			elb.Id = "elb:" + elb.LoadBalancerName
			elbs[elb.LoadBalancerName] = elb
		}
		return true
	}); err != nil {
		return nil, err
	}

	return elbs, nil

}

func (elb *ELB) Poll() []chan error {

	var errs []chan error
	elb.Stats = &ELBStats{}

	for _, m := range ELBMetrics {
		m := m
		errs = append(errs, elb.Region.Throttle.do(elb.Name+" METRICS POLL", func() error {
			return m.RunFor(elb)
		}))
	}

	return errs

}

func (elb *ELB) Inactive() bool {
	return len(elb.Instances) == 0 || (elb.Stats != nil && elb.Stats.RequestsPerSecond == 0)
}

func (elbs ELBSet) Summary() *ELBStats {

	stats := &ELBStats{}

	var has_stats bool
	for _, elb := range elbs {
		if elb.Stats != nil {
			has_stats = true
			stats.RequestsPerSecond += elb.Stats.RequestsPerSecond
			stats.HealthyHostCountAvg += elb.Stats.HealthyHostCountAvg
			stats.UnHealthyHostCountAvg += elb.Stats.UnHealthyHostCountAvg
			stats.BackendConnectionErrorsAvg += elb.Stats.BackendConnectionErrorsAvg
			stats.SurgeQueueLengthAvg += elb.Stats.SurgeQueueLengthAvg
			stats.SpilloverCountAvg += elb.Stats.SpilloverCountAvg
			if elb.Stats.Latency.Min < stats.Latency.Min || stats.Latency.Min == 0 {
				stats.Latency.Min = elb.Stats.Latency.Min
			}
			if elb.Stats.Latency.Max > stats.Latency.Max || stats.Latency.Max == 0 {
				stats.Latency.Max = elb.Stats.Latency.Max
			}
			stats.Latency.Avg += elb.Stats.Latency.Avg
			stats.StatusPerSecond.Code2XX += elb.Stats.StatusPerSecond.Code2XX
			stats.StatusPerSecond.Code3XX += elb.Stats.StatusPerSecond.Code3XX
			stats.StatusPerSecond.Code4XX += elb.Stats.StatusPerSecond.Code4XX
			stats.StatusPerSecond.Code5XX += elb.Stats.StatusPerSecond.Code5XX
		}
	}

	if !has_stats {
		return nil
	}

	stats.Latency.Min = roundTime(stats.Latency.Min)
	stats.Latency.Max = roundTime(stats.Latency.Max)
	stats.Latency.Avg = roundTime(stats.Latency.Avg / time.Duration(len(elbs)))

	return stats

}

type (
	// for easy rendering in html
	// elb
	// 	[autoscaling group]
	// 		az
	// 			[subnet]
	// 				inst

	ELBTree struct {
		AutoScalingGroups []*ELBTreeAutoScalingGroup
	}
	ELBTreeAutoScalingGroup struct {
		AutoScalingGroup  *AutoScalingGroup
		AvailabilityZones []*ELBTreeAvailabilityZone
	}
	ELBTreeAvailabilityZone struct {
		AvailabilityZone *AvailabilityZone
		Subnets          []*ELBTreeSubnet
	}
	ELBTreeSubnet struct {
		Subnet    *Subnet
		Instances []*Instance
	}

	ELBTreeAutoScalingGroupByName []*ELBTreeAutoScalingGroup
	ELBTreeAvailabilityZoneByName []*ELBTreeAvailabilityZone
	ELBTreeSubnetByName           []*ELBTreeSubnet
)

func (a ELBTreeAutoScalingGroupByName) Len() int      { return len(a) }
func (a ELBTreeAutoScalingGroupByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ELBTreeAutoScalingGroupByName) Less(i, j int) bool {
	aa, bb := a[i].AutoScalingGroup, a[j].AutoScalingGroup
	return aa != nil && bb != nil && string_less_than(aa.Name, bb.Name)
}
func (a ELBTreeAvailabilityZoneByName) Len() int      { return len(a) }
func (a ELBTreeAvailabilityZoneByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ELBTreeAvailabilityZoneByName) Less(i, j int) bool {
	aa, bb := a[i].AvailabilityZone, a[j].AvailabilityZone
	return aa != nil && bb != nil && string_less_than(aa.Name, bb.Name)
}
func (a ELBTreeSubnetByName) Len() int      { return len(a) }
func (a ELBTreeSubnetByName) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ELBTreeSubnetByName) Less(i, j int) bool {
	aa, bb := a[i].Subnet, a[j].Subnet
	return aa != nil && bb != nil && string_less_than(aa.Name, bb.Name)
}

func (elb *ELB) Tree() *ELBTree {

	tree := map[*AutoScalingGroup]map[*AvailabilityZone]map[*Subnet][]*Instance{}

	// nil is a value meaning not in this case (errr yeah...)
	for _, inst := range elb.Instances {
		if _, exists := tree[inst.AutoScalingGroup]; !exists {
			tree[inst.AutoScalingGroup] = map[*AvailabilityZone]map[*Subnet][]*Instance{}
		}
		if _, exists := tree[inst.AutoScalingGroup][inst.AvailabilityZone]; !exists {
			tree[inst.AutoScalingGroup][inst.AvailabilityZone] = map[*Subnet][]*Instance{}
		}
		tree[inst.AutoScalingGroup][inst.AvailabilityZone][inst.Subnet] = append(tree[inst.AutoScalingGroup][inst.AvailabilityZone][inst.Subnet], inst)
	}

	vetree := &ELBTree{}

	for group, azs := range tree {
		tag := &ELBTreeAutoScalingGroup{
			AutoScalingGroup: group,
		}
		for az, subnets := range azs {
			taz := &ELBTreeAvailabilityZone{
				AvailabilityZone: az,
			}
			for subnet, insts := range subnets {
				tsu := &ELBTreeSubnet{
					Subnet:    subnet,
					Instances: insts,
				}
				sort.Sort(InstanceByNameAsc(tsu.Instances))
				taz.Subnets = append(taz.Subnets, tsu)
			}
			sort.Sort(ELBTreeSubnetByName(taz.Subnets))
			tag.AvailabilityZones = append(tag.AvailabilityZones, taz)
		}
		sort.Sort(ELBTreeAvailabilityZoneByName(tag.AvailabilityZones))
		vetree.AutoScalingGroups = append(vetree.AutoScalingGroups, tag)
	}

	return vetree

}
