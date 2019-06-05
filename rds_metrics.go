package window

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

type (
	rdsmetric struct {
		name       *string
		statistics []*string
		unit       *string
		processor  func(*DBInstance, *cloudwatch.Datapoint)
	}

	DBInstanceStats struct {
		CPUUtilization      float64
		DatabaseConnections int64
		DiskQueueDepth      int64
		FreeableMemory      int64
		FreeStorageSpace    int64
		ReplicaLag          time.Duration
		SwapUsage           int64
		ReadIOPS            float64
		WriteIOPS           float64
		ReadLatency         struct {
			Min, Avg, Max time.Duration
		}
		WriteLatency struct {
			Min, Avg, Max time.Duration
		}
		DiskReadThroughputBytesPerSecond        int64
		DiskWriteThroughputBytesPerSecond       int64
		NetworkReceiveThroughputBytesPerSecond  int64
		NetworkTransmitThroughputBytesPerSecond int64
	}
)

var (
	RDSMetrics = []*rdsmetric{
		{
			name:       aws.String("CPUUtilization"),
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Percent"),
			processor: func(db *DBInstance, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					db.Stats.CPUUtilization = *point.Average
				}
			},
		},
		{
			name:       aws.String("DatabaseConnections"),
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Count"),
			processor: func(db *DBInstance, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					db.Stats.DatabaseConnections = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("DiskQueueDepth"),
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Count"),
			processor: func(db *DBInstance, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					db.Stats.DiskQueueDepth = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("FreeableMemory"),
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Bytes"),
			processor: func(db *DBInstance, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					db.Stats.FreeableMemory = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("FreeStorageSpace"),
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Bytes"),
			processor: func(db *DBInstance, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					db.Stats.FreeStorageSpace = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("ReplicaLag"),
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Seconds"),
			processor: func(db *DBInstance, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					db.Stats.ReplicaLag = time.Duration(float64(time.Second) * *point.Average)
				}
			},
		},
		{
			name:       aws.String("SwapUsage"),
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Bytes"),
			processor: func(db *DBInstance, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					db.Stats.SwapUsage = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("ReadIOPS"),
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Count/Second"),
			processor: func(db *DBInstance, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					db.Stats.ReadIOPS = *point.Average
				}
			},
		},
		{
			name:       aws.String("WriteIOPS"),
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Count/Second"),
			processor: func(db *DBInstance, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					db.Stats.WriteIOPS = *point.Average
				}
			},
		},
		{
			name:       aws.String("ReadLatency"),
			statistics: []*string{aws.String("Minimum"), aws.String("Average"), aws.String("Maximum")},
			unit:       aws.String("Seconds"),
			processor: func(db *DBInstance, point *cloudwatch.Datapoint) {
				if point.Minimum != nil {
					db.Stats.ReadLatency.Min = roundTime(time.Duration(float64(time.Second) * *point.Minimum))
				}
				if point.Average != nil {
					db.Stats.ReadLatency.Avg = roundTime(time.Duration(float64(time.Second) * *point.Average))
				}
				if point.Maximum != nil {
					db.Stats.ReadLatency.Max = roundTime(time.Duration(float64(time.Second) * *point.Maximum))
				}

			},
		},
		{
			name:       aws.String("WriteLatency"),
			statistics: []*string{aws.String("Minimum"), aws.String("Average"), aws.String("Maximum")},
			unit:       aws.String("Seconds"),
			processor: func(db *DBInstance, point *cloudwatch.Datapoint) {
				if point.Minimum != nil {
					db.Stats.WriteLatency.Min = roundTime(time.Duration(float64(time.Second) * *point.Minimum))
				}
				if point.Average != nil {
					db.Stats.WriteLatency.Avg = roundTime(time.Duration(float64(time.Second) * *point.Average))
				}
				if point.Maximum != nil {
					db.Stats.WriteLatency.Max = roundTime(time.Duration(float64(time.Second) * *point.Maximum))
				}
			},
		},
		{
			name:       aws.String("ReadThroughput"),
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Bytes/Second"),
			processor: func(db *DBInstance, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					db.Stats.DiskReadThroughputBytesPerSecond = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("WriteThroughput"),
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Bytes/Second"),
			processor: func(db *DBInstance, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					db.Stats.DiskWriteThroughputBytesPerSecond = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("NetworkReceiveThroughput"),
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Bytes/Second"),
			processor: func(db *DBInstance, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					db.Stats.NetworkReceiveThroughputBytesPerSecond = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("NetworkTransmitThroughput"),
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Bytes/Second"),
			processor: func(db *DBInstance, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					db.Stats.NetworkTransmitThroughputBytesPerSecond = int64(*point.Average)
				}
			},
		},
	}
)

func (m *rdsmetric) RunFor(db *DBInstance) error {

	resp, err := CloudWatchClient.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
		StartTime: aws.Time(time.Now().Add(-PeriodInMinutes * time.Minute)),
		EndTime:   aws.Time(time.Now()),
		Period:    aws.Int64(PeriodInMinutes * 60),
		Namespace: aws.String("AWS/RDS"),
		Dimensions: []*cloudwatch.Dimension{{
			Name:  aws.String("DBInstanceIdentifier"),
			Value: aws.String(db.DBInstanceIdentifier),
		}},
		MetricName: m.name,
		Statistics: m.statistics,
		Unit:       m.unit, // fuck this in teh face
	})
	if err == nil && len(resp.Datapoints) > 0 {
		m.processor(db, resp.Datapoints[0])
	}

	return err

}
