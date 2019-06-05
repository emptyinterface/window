package window

type (
	Classic struct {
		Region               *Region
		AvailabilityZones    []*AvailabilityZone
		SecurityGroups       []*SecurityGroup
		AutoScalingGroups    []*AutoScalingGroup
		ELBs                 []*ELB
		Instances            []*Instance
		DBInstances          []*DBInstance
		ElasticCacheClusters []*ElasticCacheCluster
		AMIs                 []*AMI
	}
)
