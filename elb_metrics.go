package window

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type (
	ELBStats struct {
		RequestsPerSecond                float64
		HealthyHostCountAvg              float64
		UnHealthyHostCountAvg            float64
		BackendConnectionErrorsAvg float64
		SurgeQueueLengthAvg              float64
		SpilloverCountAvg                float64
		Latency                          struct {
			Min, Max, Avg time.Duration
		}
		StatusPerSecond struct {
			Code2XX float64
			Code3XX float64
			Code4XX float64
			Code5XX float64
		}
	}
	elbmetric struct {
		name       *string
		statistics []*string
		unit       *string
		processor  func(*ELB, *cloudwatch.Datapoint)
	}
)

var (
	ELBMetrics = []*elbmetric{
		// {
		// 	name:       aws.String("HealthyHostCount"),
		// 	statistics: []*string{aws.String("Average")},
		// 	processor: func(elb *ELB, point *cloudwatch.Datapoint) {
		// 		if point.Average != nil {
		// 			elb.Stats.HealthyHostCountAvg = *point.Average
		// 		}
		// 	},
		// },
		{
			name:       aws.String("UnHealthyHostCount"),
			statistics: []*string{aws.String("Average")},
			processor: func(elb *ELB, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					elb.Stats.UnHealthyHostCountAvg = *point.Average
				}
			},
		},
		{
			name:       aws.String("RequestCount"),
			statistics: []*string{aws.String("Sum")},
			processor: func(elb *ELB, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					elb.Stats.RequestsPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("Latency"),
			statistics: []*string{aws.String("Minimum"), aws.String("Maximum"), aws.String("Average")},
			unit:       aws.String("Seconds"),
			processor: func(elb *ELB, point *cloudwatch.Datapoint) {
				if point.Minimum != nil {
					elb.Stats.Latency.Min = roundTime(time.Duration(*point.Minimum * float64(time.Second)))
				}
				if point.Maximum != nil {
					elb.Stats.Latency.Max = roundTime(time.Duration(*point.Maximum * float64(time.Second)))
				}
				if point.Average != nil {
					elb.Stats.Latency.Avg = roundTime(time.Duration(*point.Average * float64(time.Second)))
				}
			},
		},
		{
			name:       aws.String("HTTPCode_Backend_2XX"),
			statistics: []*string{aws.String("Sum")},
			processor: func(elb *ELB, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					elb.Stats.StatusPerSecond.Code2XX = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("HTTPCode_Backend_3XX"),
			statistics: []*string{aws.String("Sum")},
			processor: func(elb *ELB, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					elb.Stats.StatusPerSecond.Code3XX = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("HTTPCode_Backend_4XX"),
			statistics: []*string{aws.String("Sum")},
			processor: func(elb *ELB, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					elb.Stats.StatusPerSecond.Code4XX = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("HTTPCode_Backend_5XX"),
			statistics: []*string{aws.String("Sum")},
			processor: func(elb *ELB, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					elb.Stats.StatusPerSecond.Code5XX = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("BackendConnectionErrors"),
			statistics: []*string{aws.String("Average")},
			processor: func(elb *ELB, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					elb.Stats.BackendConnectionErrorsAvg = *point.Average
				}
			},
		},
		{
			name:       aws.String("SurgeQueueLength"),
			statistics: []*string{aws.String("Average")},
			processor: func(elb *ELB, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					elb.Stats.SurgeQueueLengthAvg = *point.Average
				}
			},
		},
		{
			name:       aws.String("SpilloverCount"),
			statistics: []*string{aws.String("Average")},
			processor: func(elb *ELB, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					elb.Stats.SpilloverCountAvg = *point.Average
				}
			},
		},
	}
)

func (m *elbmetric) RunFor(elb *ELB) error {

	resp, err := CloudWatchClient.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
		StartTime: aws.Time(time.Now().Add(-PeriodInMinutes * time.Minute)),
		EndTime:   aws.Time(time.Now()),
		Period:    aws.Int64(PeriodInMinutes * 60),
		Namespace: aws.String("AWS/ELB"),
		Dimensions: []*cloudwatch.Dimension{{
			Name:  aws.String("LoadBalancerName"),
			Value: aws.String(elb.Name),
		}},
		MetricName: m.name,
		Statistics: m.statistics,
		Unit:       m.unit, // fuck this in teh face
	})
	if err == nil && len(resp.Datapoints) > 0 {
		m.processor(elb, resp.Datapoints[0])
	}

	return err

}
