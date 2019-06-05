package window

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type (
	CloudWatchAlarm struct {

		// Indicates whether actions should be executed during any changes to the alarm's
		// state.
		ActionsEnabled bool

		// The list of actions to execute when this alarm transitions into an ALARM
		// state from any other state. Each action is specified as an Amazon Resource
		// Number (ARN). Currently the only actions supported are publishing to an Amazon
		// SNS topic and triggering an Auto Scaling policy.
		AlarmActions []string

		// The Amazon Resource Name (ARN) of the alarm.
		AlarmArn string

		// The time stamp of the last update to the alarm configuration. Amazon CloudWatch
		// uses Coordinated Universal Time (UTC) when returning time stamps, which do
		// not accommodate seasonal adjustments such as daylight savings time. For more
		// information, see Time stamps (http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html#about_timestamp)
		// in the Amazon CloudWatch Developer Guide.
		AlarmConfigurationUpdatedTimestamp time.Time

		// The description for the alarm.
		AlarmDescription string

		// The name of the alarm.
		AlarmName string

		// The arithmetic operation to use when comparing the specified Statistic and
		// Threshold. The specified Statistic value is used as the first operand.
		ComparisonOperator string

		// The list of dimensions associated with the alarm's associated metric.
		Dimensions []*cloudwatch.Dimension

		// The number of periods over which data is compared to the specified threshold.
		EvaluationPeriods int64

		// The list of actions to execute when this alarm transitions into an INSUFFICIENT_DATA
		// state from any other state. Each action is specified as an Amazon Resource
		// Number (ARN). Currently the only actions supported are publishing to an Amazon
		// SNS topic or triggering an Auto Scaling policy.
		//
		// The current WSDL lists this attribute as UnknownActions.
		InsufficientDataActions []string

		// The name of the alarm's metric.
		MetricName string

		// The namespace of alarm's associated metric.
		Namespace string

		// The list of actions to execute when this alarm transitions into an OK state
		// from any other state. Each action is specified as an Amazon Resource Number
		// (ARN). Currently the only actions supported are publishing to an Amazon SNS
		// topic and triggering an Auto Scaling policy.
		OKActions []string

		// The period in seconds over which the statistic is applied.
		Period int64

		// A human-readable explanation for the alarm's state.
		StateReason string

		// An explanation for the alarm's state in machine-readable JSON format
		StateReasonData string

		// The time stamp of the last update to the alarm's state. Amazon CloudWatch
		// uses Coordinated Universal Time (UTC) when returning time stamps, which do
		// not accommodate seasonal adjustments such as daylight savings time. For more
		// information, see Time stamps (http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/cloudwatch_concepts.html#about_timestamp)
		// in the Amazon CloudWatch Developer Guide.
		StateUpdatedTimestamp time.Time

		// The state value for the alarm.
		StateValue string

		// The statistic to apply to the alarm's associated metric.
		Statistic string

		// The value against which the specified statistic is compared.
		Threshold float64

		// The unit of the alarm's associated metric.
		Unit string

		Name   string
		Id     string
		State  string
		Region *Region

		AlarmActionSNSs                         []*SNSTopic
		AlarmActionAutoScalingGroups            []*AutoScalingGroup
		InsufficientDataActionSNSs              []*SNSTopic
		InsufficientDataActionAutoScalingGroups []*AutoScalingGroup
		OKActionSNSs                            []*SNSTopic
		OKActionAutoScalingGroups               []*AutoScalingGroup
	}

	CloudWatchAlarmByNameAsc []*CloudWatchAlarm
)

func (a CloudWatchAlarmByNameAsc) Len() int      { return len(a) }
func (a CloudWatchAlarmByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a CloudWatchAlarmByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func LoadCloudWatchAlarms(input *cloudwatch.DescribeAlarmsInput) (map[string]*CloudWatchAlarm, error) {

	alarms := map[string]*CloudWatchAlarm{}

	if err := CloudWatchClient.DescribeAlarmsPages(input, func(p *cloudwatch.DescribeAlarmsOutput, _ bool) bool {
		for _, ma := range p.MetricAlarms {
			alarm := &CloudWatchAlarm{
				ActionsEnabled: aws.BoolValue(ma.ActionsEnabled),
				AlarmActions:   aws.StringValueSlice(ma.AlarmActions),
				AlarmArn:       aws.StringValue(ma.AlarmArn),
				AlarmConfigurationUpdatedTimestamp: aws.TimeValue(ma.AlarmConfigurationUpdatedTimestamp),
				AlarmDescription:                   aws.StringValue(ma.AlarmDescription),
				AlarmName:                          aws.StringValue(ma.AlarmName),
				ComparisonOperator:                 aws.StringValue(ma.ComparisonOperator),
				Dimensions:                         ma.Dimensions,
				EvaluationPeriods:                  aws.Int64Value(ma.EvaluationPeriods),
				InsufficientDataActions:            aws.StringValueSlice(ma.InsufficientDataActions),
				MetricName:                         aws.StringValue(ma.MetricName),
				Namespace:                          aws.StringValue(ma.Namespace),
				OKActions:                          aws.StringValueSlice(ma.OKActions),
				Period:                             aws.Int64Value(ma.Period),
				StateReason:                        aws.StringValue(ma.StateReason),
				StateReasonData:                    aws.StringValue(ma.StateReasonData),
				StateUpdatedTimestamp:              aws.TimeValue(ma.StateUpdatedTimestamp),
				StateValue:                         aws.StringValue(ma.StateValue),
				Statistic:                          aws.StringValue(ma.Statistic),
				Threshold:                          aws.Float64Value(ma.Threshold),
				Unit:                               aws.StringValue(ma.Unit),
			}
			alarm.Name = alarm.AlarmName
			alarm.Id = "cwa:" + alarm.AlarmArn
			alarm.State = alarm.StateValue
			alarms[alarm.AlarmArn] = alarm
		}
		return true
	}); err != nil {
		return nil, err
	}

	return alarms, nil

}

func (a *CloudWatchAlarm) Inactive() bool {
	return !a.ActionsEnabled
}

func (a *CloudWatchAlarm) Summary() string {
	if i := strings.IndexByte(a.AlarmDescription, '!'); i > 0 {
		return a.AlarmDescription[:i]
	}
	if len(a.AlarmDescription) > 0 {
		return a.AlarmDescription
	}
	return fmt.Sprintf("%s/%s/%s %s %s %s for at least %s during %d evaluation period",
		a.Namespace,
		a.MetricName,
		a.Statistic,
		a.ComparisonOperator,
		strconv.FormatFloat(a.Threshold, 'f', -1, 64),
		a.Unit,
		roundTime(time.Duration(a.Period)*time.Second),
		a.EvaluationPeriods,
	)
}
