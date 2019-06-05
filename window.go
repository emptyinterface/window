package window

import (
	"fmt"
	"sort"
	"sync"

	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type (
	Tracker struct {
		counts map[string]int
		me     sync.Mutex
	}
)

const (
	PeriodInMinutes = 10
)

var (
	sess             *session.Session
	ELBClient        *elb.ELB
	EC2Client        *ec2.EC2
	ASClient         *autoscaling.AutoScaling
	ECClient         *elasticache.ElastiCache
	RDSClient        *rds.RDS
	CloudWatchClient *cloudwatch.CloudWatch
	SQSClient        *sqs.SQS
	SNSClient        *sns.SNS
	LambdaClient     *lambda.Lambda
	IAMClient        *iam.IAM

	tracker = &Tracker{
		counts: map[string]int{},
		me:     sync.Mutex{},
	}
)

func init() {
	sess = session.New(nil)
	sess.Handlers.Send.PushBack(func(req *request.Request) {
		tracker.Increment(req.ClientInfo.ServiceName)
	})
	ELBClient = elb.New(sess)
	EC2Client = ec2.New(sess)
	ASClient = autoscaling.New(sess)
	ECClient = elasticache.New(sess)
	RDSClient = rds.New(sess)
	CloudWatchClient = cloudwatch.New(sess)
	SQSClient = sqs.New(sess)
	SNSClient = sns.New(sess)
	LambdaClient = lambda.New(sess)
	IAMClient = iam.New(sess)
}

func (t *Tracker) Increment(service string) {
	t.me.Lock()
	defer t.me.Unlock()
	t.counts[service]++
}

func (t *Tracker) Report() {
	t.me.Lock()
	defer t.me.Unlock()
	names := make([]string, 0, len(t.counts))
	for name, _ := range t.counts {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Println("AWS API Call Summary")
	var total int
	for _, name := range names {
		ct := t.counts[name]
		fmt.Println(name, ct)
		total += ct
	}
	t.counts = map[string]int{}
	day := (total * (60 / 5) * 24)
	month := day * 30
	var price float64
	// first mil free
	if month > 1000000 {
		price = float64(month-1000000) / 1000 * .01 // penny per thousand
	}
	fmt.Println("Total", total, "(", day, "per day,", month, "per month )")
	fmt.Printf("%.2f/second\n", float64(day)/24/60/60)
	fmt.Printf("Est Cost $%.02f/month\n", price)

}
