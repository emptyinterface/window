package window

import (
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type (
	SQSQueue struct {
		QueueUrl                              string
		ApproximateNumberOfMessages           int64         // - returns the approximate number of visible messages in a queue. For more information, see Resources Required to Process Messages in the Amazon SQSQueue Developer Guide.
		ApproximateNumberOfMessagesNotVisible int64         // - returns the approximate number of messages that are not timed-out and not deleted. For more information, see Resources Required to Process Messages in the Amazon SQSQueue Developer Guide.
		VisibilityTimeout                     time.Duration // - returns the visibility timeout for the queue. For more information about visibility timeout, see Visibility Timeout in the Amazon SQSQueue Developer Guide.
		CreatedTimestamp                      time.Time     // - returns the time when the queue was created (epoch time in seconds).
		LastModifiedTimestamp                 time.Time     // - returns the time when the queue was last changed (epoch time in seconds).
		PolicyJSON                            string        // - returns the queue's policy.
		MaximumMessageSize                    int64         // - returns the limit of how many bytes a message can contain before Amazon SQSQueue rejects it.
		MessageRetentionPeriod                time.Duration // - returns the number of seconds Amazon SQSQueue retains a message.
		QueueArn                              string        // - returns the queue's Amazon resource name (ARN).
		ApproximateNumberOfMessagesDelayed    int64         // - returns the approximate number of messages that are pending to be added to the queue.
		DelaySeconds                          time.Duration // - returns the default delay on the queue in seconds.
		ReceiveMessageWaitTime                time.Duration // - returns the time for which a ReceiveMessage call will wait for a message to arrive.
		RedrivePolicy                         string        // - returns the parameters for dead letter queue functionality of the source queue. For more information about RedrivePolicy and dead letter queues, see Using Amazon SQSQueue Dead Letter Queues in the Amazon SQSQueue Developer Guide.

		Name             string
		Id               string
		State            string
		Region           *Region
		Policy           *SQSPolicy
		Stats            *QueueStats
		CloudWatchAlarms []*CloudWatchAlarm
	}

	SQSPolicy struct {
		ID         string                `json:"Id"`
		Statements []*SQSPolicyStatement `json:"Statement"`
		Version    string                `json:"Version"`
	}

	SQSPolicyStatement struct {
		Action        string          `json:"Action"`
		ConditionJSON json.RawMessage `json:"Condition"`
		Condition     string          `json:"-"`
		Effect        string          `json:"Effect"`
		Principal     struct {
			AWS string `json:"AWS"`
		} `json:"Principal"`
		Resource string `json:"Resource"`
		Sid      string `json:"Sid"`
	}

	SQSQueueByNameAsc []*SQSQueue
)

func (a SQSQueueByNameAsc) Len() int      { return len(a) }
func (a SQSQueueByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SQSQueueByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func LoadSQSQueues(input *sqs.ListQueuesInput) (map[string]*SQSQueue, error) {

	sqss := map[string]*SQSQueue{}

	resp, err := SQSClient.ListQueues(nil)
	if err != nil {
		return nil, err
	}

	for _, queueUrl := range resp.QueueUrls {
		if queueUrl != nil {
			s := &SQSQueue{
				QueueUrl: *queueUrl,
				Name:     filepath.Base(*queueUrl),
			}
			resp, err := SQSClient.GetQueueAttributes(&sqs.GetQueueAttributesInput{
				AttributeNames: []*string{aws.String("All")},
				QueueUrl:       queueUrl,
			})
			if err != nil {
				return nil, err
			}

			if len(resp.Attributes) == 0 {
				continue
			}
			s.ApproximateNumberOfMessages, _ = strconv.ParseInt(aws.StringValue(resp.Attributes["ApproximateNumberOfMessages"]), 10, 64)
			s.ApproximateNumberOfMessagesNotVisible, _ = strconv.ParseInt(aws.StringValue(resp.Attributes["ApproximateNumberOfMessagesNotVisible"]), 10, 64)

			n, _ := strconv.ParseInt(aws.StringValue(resp.Attributes["VisibilityTimeout"]), 10, 64)
			s.VisibilityTimeout = time.Duration(n) * time.Second

			n, _ = strconv.ParseInt(aws.StringValue(resp.Attributes["CreatedTimestamp"]), 10, 64)
			s.CreatedTimestamp = time.Unix(n, 0)

			n, _ = strconv.ParseInt(aws.StringValue(resp.Attributes["LastModifiedTimestamp"]), 10, 64)
			s.LastModifiedTimestamp = time.Unix(n, 0)

			s.PolicyJSON = aws.StringValue(resp.Attributes["Policy"])
			s.MaximumMessageSize, _ = strconv.ParseInt(aws.StringValue(resp.Attributes["MaximumMessageSize"]), 10, 64)

			n, _ = strconv.ParseInt(aws.StringValue(resp.Attributes["MessageRetentionPeriod"]), 10, 64)
			s.MessageRetentionPeriod = time.Duration(n) * time.Second

			s.QueueArn = aws.StringValue(resp.Attributes["QueueArn"])
			s.ApproximateNumberOfMessagesDelayed, _ = strconv.ParseInt(aws.StringValue(resp.Attributes["ApproximateNumberOfMessagesDelayed"]), 10, 64)

			n, _ = strconv.ParseInt(aws.StringValue(resp.Attributes["DelaySeconds"]), 10, 64)
			s.DelaySeconds = time.Duration(n) * time.Second

			n, _ = strconv.ParseInt(aws.StringValue(resp.Attributes["ReceiveMessageWaitTimeSeconds"]), 10, 64)
			s.ReceiveMessageWaitTime = time.Duration(n) * time.Second

			s.RedrivePolicy = aws.StringValue(resp.Attributes["RedrivePolicy"])

			s.Id = "sqs:" + s.QueueArn

			s.Policy = &SQSPolicy{}
			json.NewDecoder(strings.NewReader(s.PolicyJSON)).Decode(s.Policy)
			for _, statement := range s.Policy.Statements {
				statement.Condition = string(statement.ConditionJSON)
			}

			sqss[s.QueueArn] = s
		}
	}

	return sqss, nil

}

func (s *SQSQueue) Poll() []chan error {

	var errs []chan error
	s.Stats = &QueueStats{}

	for _, m := range SQSQueueMetrics {
		m := m
		errs = append(errs, s.Region.Throttle.do(s.Name+":"+*m.name+" SQSQueue METRICS POLL", func() error {
			return m.RunFor(s)
		}))
	}

	return errs

}

func (s *SQSQueue) Inactive() bool {

	if s.Stats != nil {
		return s.Stats.SentPerSecond == 0 &&
			s.Stats.EmptyReceivesPerSecond == 0 &&
			s.ApproximateNumberOfMessages == 0
	}

	return s.ApproximateNumberOfMessages == 0

}
