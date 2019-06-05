package window

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type (
	LambdaFunctionStats struct {

		// Measures the number of times a function is invoked in response to an event or invocation API call. This replaces the deprecated RequestCount metric. This includes successful and failed invocations, but does not include throttled attempts. This equals the billed requests for the function. Note that AWS Lambda only sends these metrics to CloudWatch if they have a nonzero value.
		// Units: Count
		InvocationsPerSecond float64

		// Measures the number of invocations that failed due to errors in the function (response code 4XX). This replaces the deprecated ErrorCount metric. Failed invocations may trigger a retry attempt that succeeds. This includes:
		// Handled exceptions (e.g., context.fail(error))
		// Unhandled exceptions causing the code to exit
		// Out of memory exceptions
		ErrorsPerSecond float64

		// Permissions errors
		// This does not include invocations that fail due to invocation rates exceeding default concurrent limits (error code 429) or failures due to internal service errors (error code 500).
		// Units: Count
		TimeoutsPerSecond float64

		// Measures the elapsed wall clock time from when the function code starts executing as a result of an invocation to when it stops executing. This replaces the deprecated Latency metric. The maximum data point value possible is the function timeout configuration. The billed duration will be rounded up to the nearest 100 millisecond. Note that AWS Lambda only sends these metrics to CloudWatch if they have a nonzero value.
		// Units: Milliseconds
		Duration struct {
			Min, Avg, Max time.Duration
		}

		// Measures the number of Lambda function invocation attempts that were throttled due to invocation rates exceeding the customer’s concurrent limits (error code 429). Failed invocations may trigger a retry attempt that succeeds.
		// Units: Count
		ThrottlesPerSecond float64
	}

	lfmetric struct {
		name       *string
		statistics []*string
		unit       *string
		processor  func(stats *LambdaFunctionStats, point *cloudwatch.Datapoint)
	}
)

var (
	LambdaFunctionMetrics = []*lfmetric{
		{
			name:       aws.String("Invocations"),
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *LambdaFunctionStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.InvocationsPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("Errors"),
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *LambdaFunctionStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.ErrorsPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("Timeouts"),
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *LambdaFunctionStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.TimeoutsPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("Duration"),
			statistics: []*string{aws.String("Minimum"), aws.String("Average"), aws.String("Maximum")},
			unit:       aws.String("Milliseconds"),
			processor: func(stats *LambdaFunctionStats, point *cloudwatch.Datapoint) {
				if point.Minimum != nil {
					stats.Duration.Min = roundTime(time.Duration(*point.Minimum * float64(time.Millisecond)))
				}
				if point.Maximum != nil {
					stats.Duration.Max = roundTime(time.Duration(*point.Maximum * float64(time.Millisecond)))
				}
				if point.Average != nil {
					stats.Duration.Avg = roundTime(time.Duration(*point.Average * float64(time.Millisecond)))
				}
			},
		},
		{
			name:       aws.String("Throttles"),
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *LambdaFunctionStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.ThrottlesPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
	}
)

func (m *lfmetric) RunFor(lf *LambdaFunction) error {

	resp, err := CloudWatchClient.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
		StartTime: aws.Time(time.Now().Add(-PeriodInMinutes * time.Minute)),
		EndTime:   aws.Time(time.Now()),
		Period:    aws.Int64(PeriodInMinutes * 60),
		Namespace: aws.String("AWS/Lambda"),
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String("FunctionName"),
				Value: aws.String(lf.FunctionName),
			},
		},
		MetricName: m.name,
		Statistics: m.statistics,
		Unit:       m.unit, // fuck this in teh face
	})
	if err == nil && len(resp.Datapoints) > 0 {
		m.processor(lf.Stats, resp.Datapoints[0])
	}

	return err

}
