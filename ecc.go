package window

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/emptyinterface/window/pricing"
)

type (
	ElasticCacheCluster struct {

		// This parameter is currently disabled.
		// AutoMinorVersionUpgrade bool

		// The date and time when the cache cluster was created.
		CacheClusterCreateTime time.Time

		// The user-supplied identifier of the cache cluster. This identifier is a unique
		// key that identifies a cache cluster.
		CacheClusterId string

		// The current state of this cache cluster, one of the following values: available,
		// creating, deleted, deleting, incompatible-network, modifying, rebooting cache
		// cluster nodes, restore-failed, or snapshotting.
		CacheClusterStatus string

		// The name of the compute and memory capacity node type for the cache cluster.
		//
		// Valid node types are as follows:
		//
		//  General purpose:  Current generation: cache.t2.micro, cache.t2.small, cache.t2.medium,
		// cache.m3.medium, cache.m3.large, cache.m3.xlarge, cache.m3.2xlarge Previous
		// generation: cache.t1.micro, cache.m1.small, cache.m1.medium, cache.m1.large,
		// cache.m1.xlarge  Compute optimized: cache.c1.xlarge Memory optimized  Current
		// generation: cache.r3.large, cache.r3.xlarge, cache.r3.2xlarge, cache.r3.4xlarge,
		// cache.r3.8xlarge Previous generation: cache.m2.xlarge, cache.m2.2xlarge,
		// cache.m2.4xlarge   Notes:
		//
		//  All t2 instances are created in an Amazon Virtual Private Cloud (VPC).
		// Redis backup/restore is not supported for t2 instances. Redis Append-only
		// files (AOF) functionality is not supported for t1 or t2 instances.  For a
		// complete listing of cache node types and specifications, see Amazon ElastiCache
		// Product Features and Details (http://aws.amazon.com/elasticache/details)
		// and Cache Node Type-Specific Parameters for Memcached (http://docs.aws.amazon.com/AmazonElastiCache/latest/UserGuide/CacheParameterGroups.Memcached.html#CacheParameterGroups.Memcached.NodeSpecific)
		// or Cache Node Type-Specific Parameters for Redis (http://docs.aws.amazon.com/AmazonElastiCache/latest/UserGuide/CacheParameterGroups.Redis.html#CacheParameterGroups.Redis.NodeSpecific).
		CacheNodeType string

		// A list of cache nodes that are members of the cache cluster.
		CacheNodes []*elasticache.CacheNode

		// The status of the cache parameter group.
		CacheParameterGroup *elasticache.CacheParameterGroupStatus

		// A list of cache security group elements, composed of name and status sub-elements.
		CacheSecurityGroups []*elasticache.CacheSecurityGroupMembership

		// The name of the cache subnet group associated with the cache cluster.
		CacheSubnetGroupName string

		// Represents the information required for client programs to connect to a cache
		// node.
		ConfigurationEndpoint *elasticache.Endpoint

		// The name of the cache engine (memcached or redis) to be used for this cache
		// cluster.
		Engine string

		// The version of the cache engine version that is used in this cache cluster.
		EngineVersion string

		// Describes a notification topic and its status. Notification topics are used
		// for publishing ElastiCache events to subscribers using Amazon Simple Notification
		// Service (SNS).
		NotificationConfiguration *elasticache.NotificationConfiguration

		// The number of cache nodes in the cache cluster.
		//
		// For clusters running Redis, this value must be 1. For clusters running Memcached,
		// this value must be between 1 and 20.
		NumCacheNodes int64

		// A group of settings that will be applied to the cache cluster in the future,
		// or that are currently being applied.
		PendingModifiedValues *elasticache.PendingModifiedValues

		// The name of the Availability Zone in which the cache cluster is located or
		// "Multiple" if the cache nodes are located in different Availability Zones.
		PreferredAvailabilityZone string

		// Specifies the weekly time range during which maintenance on the cache cluster
		// is performed. It is specified as a range in the format ddd:hh24:mi-ddd:hh24:mi
		// (24H Clock UTC). The minimum maintenance window is a 60 minute period. Valid
		// values for ddd are:
		//
		//  sun mon tue wed thu fri sat  Example: sun:05:00-sun:09:00
		PreferredMaintenanceWindow string

		// The replication group to which this cache cluster belongs. If this field
		// is empty, the cache cluster is not associated with any replication group.
		ReplicationGroupId string

		// A list of VPC Security Groups associated with the cache cluster.
		VPCSecurityGroups []*elasticache.SecurityGroupMembership

		// The number of days for which ElastiCache will retain automatic cache cluster
		// snapshots before deleting them. For example, if you set SnapshotRetentionLimit
		// to 5, then a snapshot that was taken today will be retained for 5 days before
		// being deleted.
		//
		// ImportantIf the value of SnapshotRetentionLimit is set to zero (0), backups
		// are turned off.
		SnapshotRetentionLimit int64

		// The daily time range (in UTC) during which ElastiCache will begin taking
		// a daily snapshot of your cache cluster.
		//
		// Example: 05:00-09:00
		SnapshotWindow string

		Name              string
		Id                string
		State             string
		Region            *Region
		AvailabilityZones []*AvailabilityZone
		Classic           *Classic
		VPC               *VPC
		SecurityGroups    []*SecurityGroup
		CloudWatchAlarms  []*CloudWatchAlarm

		Stats []*ECCNodeStats
	}

	ElasticCacheClusterByNameAsc []*ElasticCacheCluster

	ECCSet []*ElasticCacheCluster
)

func (a ElasticCacheClusterByNameAsc) Len() int      { return len(a) }
func (a ElasticCacheClusterByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ElasticCacheClusterByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func LoadCacheClusters(input *elasticache.DescribeCacheClustersInput) (map[string]*ElasticCacheCluster, error) {

	if input == nil {
		input = &elasticache.DescribeCacheClustersInput{}
	}
	input.ShowCacheNodeInfo = aws.Bool(true)

	eccs := map[string]*ElasticCacheCluster{}

	if err := ECClient.DescribeCacheClustersPages(input, func(page *elasticache.DescribeCacheClustersOutput, _ bool) bool {
		for _, cc := range page.CacheClusters {
			ecc := &ElasticCacheCluster{
				CacheClusterCreateTime:     aws.TimeValue(cc.CacheClusterCreateTime),
				CacheClusterId:             aws.StringValue(cc.CacheClusterId),
				CacheClusterStatus:         aws.StringValue(cc.CacheClusterStatus),
				CacheNodeType:              aws.StringValue(cc.CacheNodeType),
				CacheNodes:                 cc.CacheNodes,
				CacheParameterGroup:        cc.CacheParameterGroup,
				CacheSecurityGroups:        cc.CacheSecurityGroups,
				CacheSubnetGroupName:       aws.StringValue(cc.CacheSubnetGroupName),
				ConfigurationEndpoint:      cc.ConfigurationEndpoint,
				Engine:                     aws.StringValue(cc.Engine),
				EngineVersion:              aws.StringValue(cc.EngineVersion),
				NotificationConfiguration:  cc.NotificationConfiguration,
				NumCacheNodes:              aws.Int64Value(cc.NumCacheNodes),
				PendingModifiedValues:      cc.PendingModifiedValues,
				PreferredAvailabilityZone:  aws.StringValue(cc.PreferredAvailabilityZone),
				PreferredMaintenanceWindow: aws.StringValue(cc.PreferredMaintenanceWindow),
				ReplicationGroupId:         aws.StringValue(cc.ReplicationGroupId),
				VPCSecurityGroups:          cc.SecurityGroups,
				SnapshotRetentionLimit:     aws.Int64Value(cc.SnapshotRetentionLimit),
				SnapshotWindow:             aws.StringValue(cc.SnapshotWindow),
			}
			ecc.Name = ecc.CacheClusterId
			ecc.Id = "ecc:" + ecc.CacheClusterId
			ecc.State = ecc.CacheClusterStatus
			eccs[ecc.CacheClusterId] = ecc
		}
		return true
	}); err != nil {
		return nil, err
	}

	return eccs, nil

}

func (ecc *ElasticCacheCluster) priceKey() string {
	var engine string
	switch ecc.Engine {
	case "redis":
		engine = pricing.RedisCacheEngine
	case "memcached":
		engine = pricing.MemcachedCacheEngine
	}
	key := fmt.Sprintf("%s:%s:%s:%s",
		pricing.AmazonElastiCacheOfferCode,
		pricing.OnDemandTermType,
		ecc.CacheNodeType,
		engine,
	)
	return key
}

func (ecc *ElasticCacheCluster) HourlyCost() float64 {
	if offer, exists := ecc.Region.Prices[ecc.priceKey()]; exists {
		return offer.PricePerUnit
	}
	fmt.Println("miss", ecc.priceKey())
	return 0
}

func (ecc *ElasticCacheCluster) MonthlyCost() float64 {
	if offer, exists := ecc.Region.Prices[ecc.priceKey()]; exists {
		return offer.PricePerUnit * 24 * 30
	}
	fmt.Println("miss", ecc.priceKey())
	return 0
}

func (ecc *ElasticCacheCluster) Poll() []chan error {

	var errs []chan error
	ecc.Stats = nil

	for _, node := range ecc.CacheNodes {
		stats := NewECCNodeStats(ecc, node)
		ecc.Stats = append(ecc.Stats, stats)
		for _, m := range ECCMetrics {
			m := m
			errs = append(errs, ecc.Region.Throttle.do(ecc.Name+" ECC METRICS POLL", func() error {
				return m.RunFor(stats)
			}))
		}
	}

	return errs

}

func (ecc *ElasticCacheCluster) AggregateStats() *ECCNodeStats {

	stats := &ECCNodeStats{}
	if ecc.Engine == "redis" {
		stats.Redis = &ECCRedisStats{}
	} else {
		stats.Memcached = &ECCMemcachedStats{}
	}

	var redis_nodes int
	for _, stat := range ecc.Stats {
		stats.CPUUtilization += stat.CPUUtilization
		stats.FreeableMemory += stat.FreeableMemory
		stats.NetworkBytesInPerSecond += stat.NetworkBytesInPerSecond
		stats.NetworkBytesOutPerSecond += stat.NetworkBytesOutPerSecond
		stats.SwapUsage += stat.SwapUsage
		if stat.Redis != nil {
			redis_nodes++
			stats.Redis.CacheHitsPerSecond += stat.Redis.CacheHitsPerSecond
			stats.Redis.CacheMissesPerSecond += stat.Redis.CacheMissesPerSecond
			stats.Redis.CurrConnections += stat.Redis.CurrConnections
			stats.Redis.NewConnections += stat.Redis.NewConnections
			stats.Redis.CurrItems += stat.Redis.CurrItems
			stats.Redis.BytesUsedForCache += stat.Redis.BytesUsedForCache
			stats.Redis.EvictionsPerSecond += stat.Redis.EvictionsPerSecond
			stats.Redis.GetTypeCmdsPerSecond += stat.Redis.GetTypeCmdsPerSecond
			stats.Redis.SetTypeCmdsPerSecond += stat.Redis.SetTypeCmdsPerSecond
			stats.Redis.ReclaimedPerSecond += stat.Redis.ReclaimedPerSecond
			stats.Redis.ReplicationBytesPerSecond += stat.Redis.ReplicationBytesPerSecond
			if stats.Redis.ReplicationLag.Min == 0 || stat.Redis.ReplicationLag.Min < stats.Redis.ReplicationLag.Min {
				stats.Redis.ReplicationLag.Min = stat.Redis.ReplicationLag.Min
			}
			if stat.Redis.ReplicationLag.Max > stats.Redis.ReplicationLag.Max {
				stats.Redis.ReplicationLag.Max = stat.Redis.ReplicationLag.Max
			}
			stats.Redis.ReplicationLag.Avg += stat.Redis.ReplicationLag.Avg
			if stat.Redis.SaveInProgress {
				stats.Redis.SaveInProgress = true
			}
		} else if stat.Memcached != nil {
			stats.Memcached.CurrItems += stat.Memcached.CurrItems
			stats.Memcached.NewItems += stat.Memcached.NewItems
			stats.Memcached.BytesUsedForCacheItems += stat.Memcached.BytesUsedForCacheItems
			stats.Memcached.CurrConnections += stat.Memcached.CurrConnections
			stats.Memcached.NewConnections += stat.Memcached.NewConnections
			stats.Memcached.CmdFlushPerSecond += stat.Memcached.CmdFlushPerSecond
			stats.Memcached.CmdGetPerSecond += stat.Memcached.CmdGetPerSecond
			stats.Memcached.CmdSetPerSecond += stat.Memcached.CmdSetPerSecond
			stats.Memcached.EvictionsPerSecond += stat.Memcached.EvictionsPerSecond
			stats.Memcached.GetHitsPerSecond += stat.Memcached.GetHitsPerSecond
			stats.Memcached.GetMissesPerSecond += stat.Memcached.GetMissesPerSecond
			stats.Memcached.ReclaimedPerSecond += stat.Memcached.ReclaimedPerSecond
			stats.Memcached.EvictedUnfetchedPerSecond += stat.Memcached.EvictedUnfetchedPerSecond
			stats.Memcached.ExpiredUnfetchedPerSecond += stat.Memcached.ExpiredUnfetchedPerSecond
		}
	}

	stats.CPUUtilization /= float64(len(ecc.Stats))
	if stats.Redis != nil {
		stats.Redis.ReplicationLag.Avg /= time.Duration(redis_nodes)
	}

	return stats

}

func (ecc *ElasticCacheCluster) Inactive() bool {
	if ecc.CacheClusterStatus == "deleted" {
		return true
	}
	var inactive bool
	for _, stat := range ecc.Stats {
		if stat.Redis != nil {
			if stat.Redis.CurrConnections == 0 || stat.Redis.CurrItems == 0 {
				inactive = true
			}
		}
		if stat.Memcached != nil {
			if stat.Memcached.CurrConnections == 0 || stat.Memcached.CurrItems == 0 {
				inactive = true
			}
		}
	}
	return inactive
}

func (eccs ECCSet) Summary() *ECCNodeStats {

	stats := &ECCNodeStats{}

	var nodes, redis_nodes int
	for _, ecc := range eccs {
		nodes += len(ecc.Stats)
		for _, stat := range ecc.Stats {
			stats.CPUUtilization += stat.CPUUtilization
			stats.NetworkBytesInPerSecond += stat.NetworkBytesInPerSecond
			stats.NetworkBytesOutPerSecond += stat.NetworkBytesOutPerSecond
			if stat.Redis != nil {
				redis_nodes++
				if stats.Redis == nil {
					stats.Redis = &ECCRedisStats{}
				}
				stats.Redis.CacheHitsPerSecond += stat.Redis.CacheHitsPerSecond
				stats.Redis.CacheMissesPerSecond += stat.Redis.CacheMissesPerSecond
				stats.Redis.CurrConnections += stat.Redis.CurrConnections
				stats.Redis.NewConnections += stat.Redis.NewConnections
				stats.Redis.CurrItems += stat.Redis.CurrItems
				stats.Redis.BytesUsedForCache += stat.Redis.BytesUsedForCache
				stats.Redis.EvictionsPerSecond += stat.Redis.EvictionsPerSecond
				stats.Redis.GetTypeCmdsPerSecond += stat.Redis.GetTypeCmdsPerSecond
				stats.Redis.SetTypeCmdsPerSecond += stat.Redis.SetTypeCmdsPerSecond
				stats.Redis.ReclaimedPerSecond += stat.Redis.ReclaimedPerSecond
				stats.Redis.ReplicationBytesPerSecond += stat.Redis.ReplicationBytesPerSecond
				if stats.Redis.ReplicationLag.Min == 0 || stat.Redis.ReplicationLag.Min < stats.Redis.ReplicationLag.Min {
					stats.Redis.ReplicationLag.Min = stat.Redis.ReplicationLag.Min
				}
				if stat.Redis.ReplicationLag.Max > stats.Redis.ReplicationLag.Max {
					stats.Redis.ReplicationLag.Max = stat.Redis.ReplicationLag.Max
				}
				stats.Redis.ReplicationLag.Avg += stat.Redis.ReplicationLag.Avg
			} else if stat.Memcached != nil {
				if stats.Memcached == nil {
					stats.Memcached = &ECCMemcachedStats{}
				}
				stats.Memcached.CurrItems += stat.Memcached.CurrItems
				stats.Memcached.NewItems += stat.Memcached.NewItems
				stats.Memcached.BytesUsedForCacheItems += stat.Memcached.BytesUsedForCacheItems
				stats.Memcached.CurrConnections += stat.Memcached.CurrConnections
				stats.Memcached.NewConnections += stat.Memcached.NewConnections
				stats.Memcached.CmdFlushPerSecond += stat.Memcached.CmdFlushPerSecond
				stats.Memcached.CmdGetPerSecond += stat.Memcached.CmdGetPerSecond
				stats.Memcached.CmdSetPerSecond += stat.Memcached.CmdSetPerSecond
				stats.Memcached.EvictionsPerSecond += stat.Memcached.EvictionsPerSecond
				stats.Memcached.GetHitsPerSecond += stat.Memcached.GetHitsPerSecond
				stats.Memcached.GetMissesPerSecond += stat.Memcached.GetMissesPerSecond
				stats.Memcached.ReclaimedPerSecond += stat.Memcached.ReclaimedPerSecond
				stats.Memcached.EvictedUnfetchedPerSecond += stat.Memcached.EvictedUnfetchedPerSecond
				stats.Memcached.ExpiredUnfetchedPerSecond += stat.Memcached.ExpiredUnfetchedPerSecond
			}
		}
	}

	if nodes == 0 {
		return nil
	}

	stats.CPUUtilization /= float64(nodes)
	if stats.Redis != nil {
		stats.Redis.ReplicationLag.Min = roundTime(stats.Redis.ReplicationLag.Min)
		stats.Redis.ReplicationLag.Max = roundTime(stats.Redis.ReplicationLag.Max)
		stats.Redis.ReplicationLag.Avg = roundTime(stats.Redis.ReplicationLag.Avg / time.Duration(redis_nodes))
	}

	return stats

}
