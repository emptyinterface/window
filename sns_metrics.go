package window

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type (
	TopicStats struct {
		PublishedPerSecond  float64
		PublishSizeAvgBytes int64
		DeliveredPerSecond  float64
		FailedPerSecond     float64
	}

	snsmetric struct {
		name       *string
		statistics []*string
		unit       *string
		processor  func(stats *TopicStats, point *cloudwatch.Datapoint)
	}
)

var (
	SNSTopicMetrics = []*snsmetric{
		{
			name:       aws.String("NumberOfMessagesPublished"),
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *TopicStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.PublishedPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("PublishSize"),
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Bytes"),
			processor: func(stats *TopicStats, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					stats.PublishSizeAvgBytes = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("NumberOfNotificationsDelivered"),
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *TopicStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.DeliveredPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("NumberOfNotificationsFailed"),
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *TopicStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.FailedPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
	}
)

func (m *snsmetric) RunFor(t *SNSTopic) error {

	resp, err := CloudWatchClient.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
		StartTime: aws.Time(time.Now().Add(-PeriodInMinutes * time.Minute)),
		EndTime:   aws.Time(time.Now()),
		Period:    aws.Int64(PeriodInMinutes * 60),
		Namespace: aws.String("AWS/SNS"),
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String("TopicName"),
				Value: aws.String(t.TopicName),
			},
		},
		MetricName: m.name,
		Statistics: m.statistics,
		Unit:       m.unit, // fuck this in teh face
	})
	if err == nil && len(resp.Datapoints) > 0 {
		m.processor(t.Stats, resp.Datapoints[0])
	}

	return err

}
