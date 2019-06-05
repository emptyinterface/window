package window

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
)

type (
	SNSTopic struct {
		DeliveryPolicy              string
		DisplayName                 string
		EffectiveDeliveryPolicyJSON string
		Owner                       string
		PolicyJSON                  string
		SubscriptionsConfirmed      int64
		SubscriptionsDeleted        int64
		SubscriptionsPending        int64
		TopicArn                    string

		Name                      string
		Id                        string
		State                     string
		TopicName                 string // tail of arn
		Region                    *Region
		Policy                    *SNSPolicy
		EffectiveDeliveryPolicies SNSDeliveryPolicies
		Subscribers               []*SNSSubscription
		Stats                     *TopicStats
		CloudWatchAlarms          []*CloudWatchAlarm
	}

	SNSSubscription struct {
		// The subscription's endpoint (format depends on the protocol).
		Endpoint string

		// The subscription's owner.
		Owner string

		// The subscription's protocol.
		Protocol string

		// The subscription's ARN.
		SubscriptionArn string

		// The ARN of the subscription's topic.
		TopicArn string

		Name   string
		Region *Region
	}

	SNSPolicy struct {
		ID         string                `json:"Id"`
		Statements []*SNSPolicyStatement `json:"Statement"`
		Version    string                `json:"Version"`
	}

	SNSPolicyStatement struct {
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

	SNSDeliveryPolicies map[string]*SNSDeliveryPolicy
	SNSDeliveryPolicy   struct {
		DefaultHealthyRetryPolicy struct {
			BackoffFunction    string `json:"backoffFunction"`
			MaxDelayTarget     int    `json:"maxDelayTarget"`
			MinDelayTarget     int    `json:"minDelayTarget"`
			NumMaxDelayRetries int    `json:"numMaxDelayRetries"`
			NumMinDelayRetries int    `json:"numMinDelayRetries"`
			NumNoDelayRetries  int    `json:"numNoDelayRetries"`
			NumRetries         int    `json:"numRetries"`
		} `json:"defaultHealthyRetryPolicy"`
		DisableSubscriptionOverrides bool `json:"disableSubscriptionOverrides"`
	}

	SNSTopicByNameAsc        []*SNSTopic
	SNSSubscriptionByNameAsc []*SNSSubscription
)

func (a SNSTopicByNameAsc) Len() int      { return len(a) }
func (a SNSTopicByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SNSTopicByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}
func (a SNSSubscriptionByNameAsc) Len() int      { return len(a) }
func (a SNSSubscriptionByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SNSSubscriptionByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func LoadSNSTopics(input *sns.ListTopicsInput) (map[string]*SNSTopic, error) {

	snsts := map[string]*SNSTopic{}

	if err := SNSClient.ListTopicsPages(input, func(p *sns.ListTopicsOutput, _ bool) bool {
		for _, topic := range p.Topics {
			snsts[*topic.TopicArn] = &SNSTopic{TopicArn: *topic.TopicArn}
		}
		return true
	}); err != nil {
		return nil, err
	}

	errs := make(chan error, len(snsts))

	for arn, snst := range snsts {
		go func(arn string, snst *SNSTopic) {
			resp, err := SNSClient.GetTopicAttributes(&sns.GetTopicAttributesInput{TopicArn: aws.String(arn)})
			if err == nil {
				snst.DeliveryPolicy = aws.StringValue(resp.Attributes["DeliveryPolicy"])
				snst.DisplayName = aws.StringValue(resp.Attributes["DisplayName"])
				snst.EffectiveDeliveryPolicyJSON = aws.StringValue(resp.Attributes["EffectiveDeliveryPolicy"])
				snst.Owner = aws.StringValue(resp.Attributes["Owner"])
				snst.PolicyJSON = aws.StringValue(resp.Attributes["Policy"])
				snst.TopicArn = aws.StringValue(resp.Attributes["TopicArn"])
				snst.SubscriptionsConfirmed, _ = strconv.ParseInt(aws.StringValue(resp.Attributes["SubscriptionsConfirmed"]), 10, 64)
				snst.SubscriptionsDeleted, _ = strconv.ParseInt(aws.StringValue(resp.Attributes["SubscriptionsDeleted"]), 10, 64)
				snst.SubscriptionsPending, _ = strconv.ParseInt(aws.StringValue(resp.Attributes["SubscriptionsPending"]), 10, 64)
				snst.TopicName = snst.TopicArn[strings.LastIndex(snst.TopicArn, ":")+1:]
				snst.Id = "sns:" + snst.TopicArn
				if len(snst.DisplayName) > 0 {
					snst.Name = snst.DisplayName
				} else {
					snst.Name = snst.TopicName
				}
				snst.Policy = &SNSPolicy{}
				json.NewDecoder(strings.NewReader(snst.PolicyJSON)).Decode(snst.Policy)
				for _, statement := range snst.Policy.Statements {
					statement.Condition = string(statement.ConditionJSON)
				}
				snst.EffectiveDeliveryPolicies = SNSDeliveryPolicies{}
				json.NewDecoder(strings.NewReader(snst.EffectiveDeliveryPolicyJSON)).Decode(&snst.EffectiveDeliveryPolicies)
			}
			errs <- err
		}(arn, snst)
	}

	for i := 0; i < cap(errs); i++ {
		if err := <-errs; err != nil {
			return nil, err
		}
	}

	return snsts, nil

}

func LoadSNSSubscriptions(input *sns.ListSubscriptionsInput) (map[string]*SNSSubscription, error) {

	subs := map[string]*SNSSubscription{}

	if err := SNSClient.ListSubscriptionsPages(input, func(p *sns.ListSubscriptionsOutput, _ bool) bool {
		for _, sub := range p.Subscriptions {
			ss := &SNSSubscription{
				Endpoint:        aws.StringValue(sub.Endpoint),
				Owner:           aws.StringValue(sub.Owner),
				Protocol:        aws.StringValue(sub.Protocol),
				SubscriptionArn: aws.StringValue(sub.SubscriptionArn),
				TopicArn:        aws.StringValue(sub.TopicArn),
			}
			ss.Name = ss.Endpoint
			subs[ss.SubscriptionArn] = ss
		}
		return true
	}); err != nil {
		return nil, err
	}

	return subs, nil

}

func (t *SNSTopic) Poll() []chan error {

	var errs []chan error
	t.Stats = &TopicStats{}

	for _, m := range SNSTopicMetrics {
		m := m
		errs = append(errs, t.Region.Throttle.do(t.Name+":"+*m.name+" SNSTopic METRICS POLL", func() error {
			return m.RunFor(t)
		}))
	}

	return errs

}

func (s *SNSSubscription) TargetName() string {

	switch strings.ToLower(s.Protocol) {
	case "http", "https":
		u, err := url.Parse(s.Endpoint)
		if err == nil {
			return u.Host
		}
		return s.Endpoint
	case "email", "email-json", "sms":
		return s.Endpoint
	case "sqs":
		for _, q := range s.Region.SQSQueues {
			if q.QueueArn == s.Endpoint {
				return q.Name
			}
		}
		if i := strings.LastIndexByte(s.Endpoint, ':'); i > -1 {
			return "ext://" + s.Endpoint[i+1:]
		}
		return "ext://" + s.Endpoint
	}
	return s.Endpoint

}

func (t *SNSTopic) Inactive() bool {
	return len(t.Subscribers) == 0 ||
		(t.Stats != nil && t.Stats.PublishedPerSecond == 0)
}
