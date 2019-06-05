package window

import (
	"time"

	"github.com/aws/aws-sdk-go/service/lambda"

	"github.com/aws/aws-sdk-go/aws"
)

type (
	LambdaFunction struct {

		// It is the SHA256 hash of your function deployment package.
		CodeSha256 string

		// The size, in bytes, of the function .zip file you uploaded.
		CodeSize int64

		// The user-provided description.
		Description string

		// The Amazon Resource Name (ARN) assigned to the function.
		FunctionArn string

		// The name of the function.
		FunctionName string

		// The function Lambda calls to begin executing your function.
		Handler string

		// The timestamp of the last time you updated the function.
		LastModified string

		// The memory size, in MB, you configured for the function. Must be a multiple
		// of 64 MB.
		MemorySize int64

		// The Amazon Resource Name (ARN) of the IAM role that Lambda assumes when it
		// executes your function to access any other Amazon Web Services (AWS) resources.
		Role string

		// The runtime environment for the Lambda function.
		Runtime string

		// The function execution time at which Lambda should terminate the function.
		// Because the execution time has cost implications, we recommend you set this
		// value based on your expected execution time. The default is 3 seconds.
		Timeout time.Duration

		// The version of the Lambda function.
		Version string

		// VPC configuration associated with your Lambda function.
		VpcConfig *lambda.VpcConfigResponse

		Name             string
		Id               string
		State            string
		LastModifiedTime time.Time
		Region           *Region
		VPC              *VPC
		SecurityGroups   []*SecurityGroup
		Subnets          []*Subnet
		CloudWatchAlarms []*CloudWatchAlarm
		Stats            *LambdaFunctionStats
	}

	LambdaFunctionsByNameAsc []*LambdaFunction
)

func (a LambdaFunctionsByNameAsc) Len() int      { return len(a) }
func (a LambdaFunctionsByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a LambdaFunctionsByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func LoadLambdaFunctions(input *lambda.ListFunctionsInput) (map[string]*LambdaFunction, error) {

	funcs := map[string]*LambdaFunction{}

	if err := LambdaClient.ListFunctionsPages(input, func(p *lambda.ListFunctionsOutput, _ bool) bool {
		for _, f := range p.Functions {
			lf := &LambdaFunction{
				CodeSha256:   aws.StringValue(f.CodeSha256),
				CodeSize:     aws.Int64Value(f.CodeSize),
				Description:  aws.StringValue(f.Description),
				FunctionArn:  aws.StringValue(f.FunctionArn),
				FunctionName: aws.StringValue(f.FunctionName),
				Handler:      aws.StringValue(f.Handler),
				LastModified: aws.StringValue(f.LastModified),
				MemorySize:   aws.Int64Value(f.MemorySize),
				Role:         aws.StringValue(f.Role),
				Runtime:      aws.StringValue(f.Runtime),
				Timeout:      time.Duration(aws.Int64Value(f.Timeout)) * time.Second,
				Version:      aws.StringValue(f.Version),
				VpcConfig:    f.VpcConfig,
			}
			lf.Name = lf.FunctionName
			lf.Id = "lambda:" + lf.FunctionArn
			lf.LastModifiedTime, _ = time.Parse(`2006-01-02T15:04:05.999999999-0700`, lf.LastModified) // fuck
			funcs[lf.FunctionName] = lf
		}

		return true
	}); err != nil {
		return nil, err
	}

	return funcs, nil

}

func (lf *LambdaFunction) Poll() []chan error {

	var errs []chan error
	lf.Stats = &LambdaFunctionStats{}

	for _, m := range LambdaFunctionMetrics {
		m := m
		errs = append(errs, lf.Region.Throttle.do(lf.Name+":"+*m.name+" Lambda METRICS POLL", func() error {
			return m.RunFor(lf)
		}))
	}

	return errs

}

func (lf *LambdaFunction) Inactive() bool {
	if lf.Stats == nil {
		return false // default to active if we can't tell
	}
	return lf.Stats.InvocationsPerSecond == 0
}
