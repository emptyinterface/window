package window

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/elasticache"
)

type (
	ECCNodeStats struct {
		Cluster *ElasticCacheCluster
		Node    *elasticache.CacheNode

		// host
		CPUUtilization           float64
		FreeableMemory           int64
		NetworkBytesInPerSecond  int64
		NetworkBytesOutPerSecond int64
		SwapUsage                int64

		Memcached *ECCMemcachedStats
		Redis     *ECCRedisStats
	}

	ECCMemcachedStats struct {
		CurrItems                 int64   // (Count) A count of the number of items currently stored in the cache.	Count
		NewItems                  int64   // (Count) The number of new items the cache has stored. This is derived from the memcached total_items statistic by recording the change in total_items across a period of time.	Count
		BytesUsedForCacheItems    int64   // (Bytes) The number of bytes used to store cache items.	Bytes
		CurrConnections           int64   // (Count) A count of the number of connections connected to the cache at an instant in time.	Count
		NewConnections            int64   // (Count) The number of new connections the cache has received. This is derived from the memcached total_connections statistic by recording the change in total_connections across a period of time. This will always be at least 1, due to a connection reserved for a ElastiCache.	Count
		CmdFlushPerSecond         float64 // (Count) The number of flush commands the cache has received.	Count
		CmdGetPerSecond           float64 // (Count) The number of get commands the cache has received.	Count
		CmdSetPerSecond           float64 // (Count) The number of set commands the cache has received.	Count
		EvictionsPerSecond        float64 // (Count) The number of non-expired items the cache evicted to allow space for new writes.	Count
		GetHitsPerSecond          float64 // (Count) The number of get requests the cache has received where the key requested was found.	Count
		GetMissesPerSecond        float64 // (Count) The number of get requests the cache has received where the key requested was not found.	Count
		ReclaimedPerSecond        float64 // (Count) The number of expired items the cache evicted to allow space for new writes.	Count
		EvictedUnfetchedPerSecond float64 // (Count) The number of valid items evicted from the least recently used cache (LRU) which were never touched after being set.	Count
		ExpiredUnfetchedPerSecond float64 // (Count) The number of expired items reclaimed from the LRU which were never touched after being set.	Count
	}

	ECCRedisStats struct {
		// redis
		CacheHitsPerSecond        float64 // (Count) The number of successful key lookups.	Count
		CacheMissesPerSecond      float64 // (Count) The number of unsuccessful key lookups.	Count
		CurrConnections           int64   // (Count) The number of client connections, excluding connections from read replicas.	Count
		NewConnections            int64   // (Count) The total number of connections that have been accepted by the server during this period.	Count
		CurrItems                 int64   // (Count) The number of items in the cache. This is derived from the Redis keyspace statistic, summing all of the keys in the entire keyspace.	Count
		BytesUsedForCache         int64   // (Bytes) The total number of bytes allocated by Redis.	Bytes
		EvictionsPerSecond        float64 // (Count) The number of keys that have been evicted due to the maxmemory limit.	Count
		GetTypeCmdsPerSecond      float64 // (Count) The total number of get types of commands. This is derived from the Redis commandstats statistic by summing all of the get types of commands (get, mget, hget, etc.)	Count
		SetTypeCmdsPerSecond      float64 // (Count) The total number of set types of commands. This is derived from the Redis commandstats statistic by summing all of the set types of commands (set, hset, etc.)	Count
		ReclaimedPerSecond        float64 // (Count) The total number of key expiration events.	Count
		ReplicationBytesPerSecond int64   // (Bytes) For primaries with attached replicas, ReplicationBytes reports the number of bytes that the primary is sending to all of its replicas. This metric is representative of the write load on the replication group. For replicas and standalone primaries, ReplicationBytes is always 0.	Bytes
		ReplicationLag            struct {
			Min, Max, Avg time.Duration
		} // (Seconds) This metric is only applicable for a cache node running as a read replica. It represents how far behind, in seconds, the replica is in applying changes from the primary cache cluster.	Seconds
		SaveInProgress bool // (Count) This binary metric returns 1 whenever a background save (forked or forkless) is in progress, and 0 otherwise. A background save process is typically used during snapshots and syncs. These operations can cause degraded performance. Using the SaveInProgress metric, you can diagnose whether or not degraded performance was caused by a background save process.	Count
	}

	eccmetric struct {
		name       *string
		engine     string
		statistics []*string
		unit       *string
		processor  func(*ECCNodeStats, *cloudwatch.Datapoint)
	}
)

const (
	ECCEngineMemcached = "memcached"
	ECCEngineRedis     = "redis"
)

var (
	ECCMetrics = []*eccmetric{

		// host
		{
			name:       aws.String("CPUUtilization"),
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Percent"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					stats.CPUUtilization = *point.Average
				}
			},
		},
		{
			name:       aws.String("FreeableMemory"),
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Bytes"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					stats.FreeableMemory = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("NetworkBytesIn"),
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Bytes"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.NetworkBytesInPerSecond = int64(*point.Sum / PeriodInMinutes / 60)
				}
			},
		},
		{
			name:       aws.String("NetworkBytesOut"),
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Bytes"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.NetworkBytesOutPerSecond = int64(*point.Sum / PeriodInMinutes / 60)
				}
			},
		},
		{
			name:       aws.String("SwapUsage"),
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Bytes"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					stats.SwapUsage = int64(*point.Average)
				}
			},
		},

		// memcached
		{
			name:       aws.String("CurrItems"),
			engine:     ECCEngineMemcached,
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					stats.Memcached.CurrItems = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("NewItems"),
			engine:     ECCEngineMemcached,
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					stats.Memcached.NewItems = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("BytesUsedForCacheItems"),
			engine:     ECCEngineMemcached,
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Bytes"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					stats.Memcached.BytesUsedForCacheItems = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("CurrConnections"),
			engine:     ECCEngineMemcached,
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					stats.Memcached.CurrConnections = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("NewConnections"),
			engine:     ECCEngineMemcached,
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					stats.Memcached.NewConnections = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("CmdFlush"),
			engine:     ECCEngineMemcached,
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.Memcached.CmdFlushPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("CmdGet"),
			engine:     ECCEngineMemcached,
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.Memcached.CmdGetPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("CmdSet"),
			engine:     ECCEngineMemcached,
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.Memcached.CmdSetPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("Evictions"),
			engine:     ECCEngineMemcached,
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.Memcached.EvictionsPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("GetHits"),
			engine:     ECCEngineMemcached,
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.Memcached.GetHitsPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("GetMisses"),
			engine:     ECCEngineMemcached,
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.Memcached.GetMissesPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("Reclaimed"),
			engine:     ECCEngineMemcached,
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.Memcached.ReclaimedPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("EvictedUnfetched"),
			engine:     ECCEngineMemcached,
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.Memcached.EvictedUnfetchedPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("ExpiredUnfetched"),
			engine:     ECCEngineMemcached,
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.Memcached.ExpiredUnfetchedPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},

		// redis
		{
			name:       aws.String("CacheHits"),
			engine:     ECCEngineRedis,
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.Redis.CacheHitsPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("CacheMisses"),
			engine:     ECCEngineRedis,
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.Redis.CacheMissesPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("CurrConnections"),
			engine:     ECCEngineRedis,
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					stats.Redis.CurrConnections = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("NewConnections"),
			engine:     ECCEngineRedis,
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					stats.Redis.NewConnections = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("CurrItems"),
			engine:     ECCEngineRedis,
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					stats.Redis.CurrItems = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("BytesUsedForCache"),
			engine:     ECCEngineRedis,
			statistics: []*string{aws.String("Average")},
			unit:       aws.String("Bytes"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Average != nil {
					stats.Redis.BytesUsedForCache = int64(*point.Average)
				}
			},
		},
		{
			name:       aws.String("Evictions"),
			engine:     ECCEngineRedis,
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.Redis.EvictionsPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("GetTypeCmds"),
			engine:     ECCEngineRedis,
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.Redis.GetTypeCmdsPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("SetTypeCmds"),
			engine:     ECCEngineRedis,
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.Redis.SetTypeCmdsPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("Reclaimed"),
			engine:     ECCEngineRedis,
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.Redis.ReclaimedPerSecond = *point.Sum / PeriodInMinutes / 60
				}
			},
		},
		{
			name:       aws.String("ReplicationBytes"),
			engine:     ECCEngineRedis,
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Bytes"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil {
					stats.Redis.ReplicationBytesPerSecond = int64(*point.Sum / PeriodInMinutes / 60)
				}
			},
		},
		{
			name:       aws.String("ReplicationLag"),
			engine:     ECCEngineRedis,
			statistics: []*string{aws.String("Minimum"), aws.String("Maximum"), aws.String("Average")},
			unit:       aws.String("Seconds"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Minimum != nil {
					stats.Redis.ReplicationLag.Min = roundTime(time.Duration(*point.Minimum * float64(time.Second)))
				}
				if point.Maximum != nil {
					stats.Redis.ReplicationLag.Max = roundTime(time.Duration(*point.Maximum * float64(time.Second)))
				}
				if point.Average != nil {
					stats.Redis.ReplicationLag.Avg = roundTime(time.Duration(*point.Average * float64(time.Second)))
				}
			},
		},
		{
			name:       aws.String("SaveInProgress"),
			engine:     ECCEngineRedis,
			statistics: []*string{aws.String("Sum")},
			unit:       aws.String("Count"),
			processor: func(stats *ECCNodeStats, point *cloudwatch.Datapoint) {
				if point.Sum != nil && *point.Sum > 0 {
					stats.Redis.SaveInProgress = true
				}
			},
		},
	}
)

func NewECCNodeStats(ecc *ElasticCacheCluster, node *elasticache.CacheNode) *ECCNodeStats {
	stats := &ECCNodeStats{
		Cluster: ecc,
		Node:    node,
	}
	switch ecc.Engine {
	case ECCEngineMemcached:
		stats.Memcached = &ECCMemcachedStats{}
	case ECCEngineRedis:
		stats.Redis = &ECCRedisStats{}
	}
	return stats
}

func (m *eccmetric) RunFor(stats *ECCNodeStats) error {

	if len(m.engine) > 0 && m.engine != stats.Cluster.Engine {
		return nil
	}

	resp, err := CloudWatchClient.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
		StartTime: aws.Time(time.Now().Add(-PeriodInMinutes * time.Minute)),
		EndTime:   aws.Time(time.Now()),
		Period:    aws.Int64(PeriodInMinutes * 60),
		Namespace: aws.String("AWS/ElastiCache"),
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String("CacheClusterId"),
				Value: aws.String(stats.Cluster.CacheClusterId),
			},
			{
				Name:  aws.String("CacheNodeId"),
				Value: stats.Node.CacheNodeId,
			},
		},
		MetricName: m.name,
		Statistics: m.statistics,
		Unit:       m.unit, // fuck this in teh face
	})
	if err == nil && len(resp.Datapoints) > 0 {
		m.processor(stats, resp.Datapoints[0])
	}

	return err

}
