package window

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type (
	QueueStats struct {

		// The number of messages added to a queue.
		SentPerSecond float64

		// The size of messages added to a queue.
		MessageSizeAvgBytes int64

		// The number of messages returned by calls to the ReceiveMessage API action.
		ReceivedPerSecond float64

		// The number of ReceiveMessage API calls that did not return a message.
		EmptyReceivesPerSecond float64

		// The number of messages deleted from the queue.
		DeletedPerSecond float64
	}

	sqsmetric struct {
		name       *string
		statistics []*string
		unit       *string
		processor  func(*QueueStats, *cloudwatch.Datapoint)
	}
)

var (
	SQSQueueMetrics = []*sqsmetric{
		{
			name:       aws.String("NumberOfMessagesSent"),
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *QueueStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.SentPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("SentMessageSize"),
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Bytes"),
			processor: func(stats *QueueStats, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					stats.MessageSizeAvgBytes = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("NumberOfMessagesReceived"),
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *QueueStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.ReceivedPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("NumberOfEmptyReceives"),
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *QueueStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.EmptyReceivesPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("NumberOfMessagesDeleted"),
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *QueueStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.DeletedPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
	}
)

func (m *sqsmetric) RunFor(s *SQSQueue) error {

	resp, err := CloudWatchClient.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
		StartTime: aws.Time(time.Now().Add(-PeriodInMinutes * time.Minute)),
		EndTime:   aws.Time(time.Now()),
		Period:    aws.Int64(PeriodInMinutes * 60),
		Namespace: aws.String("AWS/SQS"),
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String("QueueName"),
				Value: aws.String(s.Name),
			},
		},
		MetricName: m.name,
		Statistics: m.statistics,
		Unit:       m.unit, // fuck this in teh face
	})
	if err == nil && len(resp.Datapoints) > 0 {
		m.processor(s.Stats, resp.Datapoints[0])
	}

	return err

}
