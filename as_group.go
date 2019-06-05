package window

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
)

type (
	AutoScalingGroup struct {
		// The Amazon Resource Name (ARN) of the group.
		AutoScalingGroupARN string

		// The name of the group.
		AutoScalingGroupName string

		// One or more Availability Zones for the group.
		AvailabilityZoneNames []string

		// The date and time the group was created.
		CreatedTime time.Time

		// The number of seconds after a scaling activity completes before any further
		// scaling activities can start.
		DefaultCooldown int64

		// The desired size of the group.
		DesiredCapacity int64

		// The metrics enabled for the group.
		EnabledMetrics []*autoscaling.EnabledMetric

		// The amount of time that Auto Scaling waits before checking an instance's
		// health status. The grace period begins when an instance comes into service.
		HealthCheckGracePeriod int64

		// The service of interest for the health status check, which can be either
		// EC2 for Amazon EC2 or ELB for Elastic Load Balancing.
		HealthCheckType string

		// The EC2 instances associated with the group.
		AutoScalingInstances []*autoscaling.Instance

		// The name of the associated launch configuration.
		LaunchConfigurationName string

		// One or more load balancers associated with the group.
		LoadBalancerNames []string

		// The maximum size of the group.
		MaxSize int64

		// The minimum size of the group.
		MinSize int64

		// The name of the placement group into which you'll launch your instances,
		// if any. For more information, see Placement Groups (http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/placement-groups.html).
		PlacementGroup string

		// The current state of the group when DeleteAutoScalingGroup is in progress.
		Status string

		// The suspended processes associated with the group.
		SuspendedProcesses []*autoscaling.SuspendedProcess

		// The tags for the group.
		Tags []*autoscaling.TagDescription

		// The termination policies for the group.
		TerminationPolicies []string

		// One or more subnet IDs, if applicable, separated by commas.
		//
		// If you specify VPCZoneIdentifier and AvailabilityZones, ensure that the
		// Availability Zones of the subnets match the values for AvailabilityZones.
		VPCZoneIdentifier string

		Name              string
		Id                string
		State             string
		AvailabilityZones []*AvailabilityZone
		Instances         []*Instance
		CloudWatchAlarms  []*CloudWatchAlarm
	}

	AutoScalingGroupByNameAsc []*AutoScalingGroup
)

func (a AutoScalingGroupByNameAsc) Len() int      { return len(a) }
func (a AutoScalingGroupByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a AutoScalingGroupByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func LoadAutoScalingGroups(input *autoscaling.DescribeAutoScalingGroupsInput) (map[string]*AutoScalingGroup, error) {

	as_groups := map[string]*AutoScalingGroup{}

	if err := ASClient.DescribeAutoScalingGroupsPages(input, func(page *autoscaling.DescribeAutoScalingGroupsOutput, _ bool) bool {
		for _, asgroup := range page.AutoScalingGroups {
			// Instances:               aws.[]Value(asgroup.Instances),               // []*Instance
			as := &AutoScalingGroup{
				AutoScalingGroupARN:     aws.StringValue(asgroup.AutoScalingGroupARN),
				AutoScalingGroupName:    aws.StringValue(asgroup.AutoScalingGroupName),
				AvailabilityZoneNames:   aws.StringValueSlice(asgroup.AvailabilityZones),
				CreatedTime:             aws.TimeValue(asgroup.CreatedTime),
				DefaultCooldown:         aws.Int64Value(asgroup.DefaultCooldown),
				DesiredCapacity:         aws.Int64Value(asgroup.DesiredCapacity),
				EnabledMetrics:          asgroup.EnabledMetrics,
				HealthCheckGracePeriod:  aws.Int64Value(asgroup.HealthCheckGracePeriod),
				HealthCheckType:         aws.StringValue(asgroup.HealthCheckType),
				AutoScalingInstances:    asgroup.Instances,
				LaunchConfigurationName: aws.StringValue(asgroup.LaunchConfigurationName),
				LoadBalancerNames:       aws.StringValueSlice(asgroup.LoadBalancerNames),
				MaxSize:                 aws.Int64Value(asgroup.MaxSize),
				MinSize:                 aws.Int64Value(asgroup.MinSize),
				PlacementGroup:          aws.StringValue(asgroup.PlacementGroup),
				Status:                  aws.StringValue(asgroup.Status),
				SuspendedProcesses:      asgroup.SuspendedProcesses,
				Tags:                    asgroup.Tags,
				TerminationPolicies:     aws.StringValueSlice(asgroup.TerminationPolicies),
				VPCZoneIdentifier:       aws.StringValue(asgroup.VPCZoneIdentifier),
			}
			as.Name = TagDescriptionOrDefault(as.Tags, "Name", as.AutoScalingGroupName)
			as.Id = "asg:" + as.AutoScalingGroupARN
			as.State = as.Status
			as_groups[as.AutoScalingGroupName] = as
		}
		return true
	}); err != nil {
		return nil, err
	}

	return as_groups, nil

}

func (asg *AutoScalingGroup) Inactive() bool {
	return len(asg.Instances) == 0
}
