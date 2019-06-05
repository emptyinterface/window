package window

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/emptyinterface/window/pricing"
)

type (
	Region struct {
		*sync.Mutex

		Name    string
		Classic *Classic
		VPCs    []*VPC
		Prices  map[string]*pricing.Row

		InternetGateways []*InternetGateway
		CustomerGateways []*CustomerGateway
		VPGateways       []*VPGateway
		VPNConnections   []*VPNConnection

		DBInstances          DBInstanceSet
		ElasticCacheClusters ECCSet
		ELBs                 ELBSet

		AvailabilityZones []*AvailabilityZone
		Instances         []*Instance
		AMIs              []*AMI
		AutoScalingGroups []*AutoScalingGroup

		SQSQueues        []*SQSQueue
		SNSTopics        []*SNSTopic
		SNSSubscriptions []*SNSSubscription
		LambdaFunctions  []*LambdaFunction
		CloudWatchAlarms []*CloudWatchAlarm
		SecurityGroups   []*SecurityGroup

		// location of pem files corresponding to ec2 key names
		sshKeyPath string

		Items map[string]interface{}

		Throttle *throttle
	}
)

func NewRegion(name string) *Region {
	r := &Region{}
	r.Mutex = &sync.Mutex{}
	r.Name = name
	r.Classic = &Classic{}
	r.Throttle = NewThrottle(1, 100, time.Second)
	r.Prices = map[string]*pricing.Row{}
	r.Items = map[string]interface{}{}

	if table, err := pricing.LoadTable(); err == nil {
		for _, row := range table.Rows {
			if row.Region != r.Name {
				continue
			}
			switch row.OfferCode {
			case pricing.AmazonEC2OfferCode:
				if len(row.InstanceType) > 0 {
					key := fmt.Sprintf("%s:%s:%s:%s:%s",
						row.OfferCode,
						row.TermType,
						row.Tenancy,
						row.InstanceType,
						row.OperatingSystem,
					)
					r.Prices[key] = row
				}
			case pricing.AmazonRDSOfferCode:
				if len(row.InstanceType) > 0 {
					key := fmt.Sprintf("%s:%s:%s:%s:%s",
						row.OfferCode,
						row.TermType,
						row.DeploymentOption,
						row.InstanceType,
						row.DatabaseEngine,
					)
					r.Prices[key] = row
				}
			case pricing.AmazonElastiCacheOfferCode:
				if len(row.InstanceType) > 0 {
					key := fmt.Sprintf("%s:%s:%s:%s",
						row.OfferCode,
						row.TermType,
						row.InstanceType,
						row.CacheEngine,
					)
					r.Prices[key] = row
				}
			}
		}
	}

	return r
}

func (region *Region) SetSSHKeyPath(path string) {
	region.sshKeyPath = path
}

func (region *Region) Refresh() error {

	var (
		vpcs                    map[string]*VPC
		security_groups         map[string]*SecurityGroup
		acls                    map[string]*ACL
		route_tables            map[string]*RouteTable
		subnets                 map[string]*Subnet
		elbs                    map[string]*ELB
		instances               map[string]*Instance
		availability_zones      map[string]*AvailabilityZone
		internet_gateways       map[string]*InternetGateway
		customer_gateways       map[string]*CustomerGateway
		vp_gateways             map[string]*VPGateway
		vpn_connections         map[string]*VPNConnection
		vpc_endpoints           map[string]*VPCEndpoint
		vpc_peering_connections map[string]*VPCPeeringConnection
		as_groups               map[string]*AutoScalingGroup
		amis                    map[string]*AMI
		db_instances            map[string]*DBInstance
		ec_clusters             map[string]*ElasticCacheCluster
		sqs_queues              map[string]*SQSQueue
		sns_topics              map[string]*SNSTopic
		sns_subscribers         map[string]*SNSSubscription
		cloudwatch_alarms       map[string]*CloudWatchAlarm
		lambda_functions        map[string]*LambdaFunction
		enis                    map[string]*ENI
		nat_gateways            map[string]*NATGateway

		errs []chan error
	)

	errs = append(errs, region.Throttle.do("LoadInstances", func() (err error) {
		instances, err = LoadInstances(nil)
		if err != nil {
			return
		}
		imageIds := map[string]struct{}{}
		for _, inst := range instances {
			imageIds[inst.ImageId] = struct{}{}
		}
		if len(imageIds) > 0 {
			input := &ec2.DescribeImagesInput{}
			for imageId, _ := range imageIds {
				input.ImageIds = append(input.ImageIds, aws.String(imageId))
			}
			amis, err = LoadAMIs(input)
		} else {
			amis = map[string]*AMI{}
		}
		return
	}))
	errs = append(errs, region.Throttle.do("LoadCacheClusters", func() (err error) {
		ec_clusters, err = LoadCacheClusters(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadDBInstances", func() (err error) {
		db_instances, err = LoadDBInstances(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadSecurityGroups", func() (err error) {
		security_groups, err = LoadSecurityGroups(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadVPCs", func() (err error) {
		vpcs, err = LoadVPCs(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadInternetGateways", func() (err error) {
		internet_gateways, err = LoadInternetGateways(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadCustomerGateways", func() (err error) {
		customer_gateways, err = LoadCustomerGateways(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadVPGateways", func() (err error) {
		vp_gateways, err = LoadVPGateways(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadVPNConnections", func() (err error) {
		vpn_connections, err = LoadVPNConnections(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadAvailabilityZones", func() (err error) {
		availability_zones, err = LoadAvailabilityZones(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadACLs", func() (err error) {
		acls, err = LoadACLs(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadRouteTables", func() (err error) {
		route_tables, err = LoadRouteTables(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadSubnets", func() (err error) {
		subnets, err = LoadSubnets(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadELBs", func() (err error) {
		elbs, err = LoadELBs(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadAutoScalingGroups", func() (err error) {
		as_groups, err = LoadAutoScalingGroups(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadVPCEndpoints", func() (err error) {
		vpc_endpoints, err = LoadVPCEndpoints(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadVPCPeeringConnections", func() (err error) {
		vpc_peering_connections, err = LoadVPCPeeringConnections(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadSQSQueues", func() (err error) {
		sqs_queues, err = LoadSQSQueues(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadSNSTopics", func() (err error) {
		sns_topics, err = LoadSNSTopics(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadSNSSubscriptions", func() (err error) {
		sns_subscribers, err = LoadSNSSubscriptions(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadCloudWatchAlarms", func() (err error) {
		cloudwatch_alarms, err = LoadCloudWatchAlarms(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadLambdaFunctions", func() (err error) {
		lambda_functions, err = LoadLambdaFunctions(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadENIs", func() (err error) {
		enis, err = LoadENIs(nil)
		return
	}))
	errs = append(errs, region.Throttle.do("LoadNATGateways", func() (err error) {
		nat_gateways, err = LoadNATGateways(nil)
		return
	}))

	for _, errchan := range errs {
		if err := <-errchan; err != nil {
			return err
		}
	}

	start := time.Now()

	// store address of existing region
	prev_region := region

	// create a new region to populate
	region = NewRegion(prev_region.Name)

	// disable the new Throttle and use the previous
	region.Throttle.stop()
	region.Throttle = prev_region.Throttle
	region.Prices = prev_region.Prices
	region.Mutex = prev_region.Mutex
	region.SetSSHKeyPath(prev_region.sshKeyPath)

	// swap the new region data into

	// thi sisso gross
	defer func() {
		region.Lock() // using shared mutex
		defer region.Unlock()
		*prev_region = *region // copy new region in
	}()

	for _, inst := range instances {

		inst.Region = region
		region.Instances = append(region.Instances, inst)

		if vpc, exists := vpcs[inst.VpcId]; exists {
			vpc.Instances = append(vpc.Instances, inst)
			inst.VPC = vpc
		} else {
			region.Classic.Instances = append(region.Classic.Instances, inst)
			inst.Classic = region.Classic
		}

		if inst.Placement != nil && inst.Placement.AvailabilityZone != nil {
			if az, exists := availability_zones[*inst.Placement.AvailabilityZone]; exists {
				az.Instances = append(az.Instances, inst)
				inst.AvailabilityZone = az
			}
		}

		if subnet, exists := subnets[inst.SubnetId]; exists {
			subnet.Instances = append(subnet.Instances, inst)
			inst.Subnet = subnet
		}

		for _, sgname := range inst.SecurityGroupNames {
			if sg, exists := security_groups[*sgname.GroupId]; exists {
				inst.SecurityGroups = append(inst.SecurityGroups, sg)
				sg.Instances = append(sg.Instances, inst)
			}
		}

		if ami, exists := amis[inst.ImageId]; exists {
			ami.Instances = append(ami.Instances, inst)
			inst.AMI = ami
		}

	}

	for _, elb := range elbs {

		region.ELBs = append(region.ELBs, elb)
		elb.Region = region

		if vpc, exists := vpcs[elb.VpcId]; exists {
			vpc.ELBs = append(vpc.ELBs, elb)
			elb.VPC = vpc
		} else {
			region.Classic.ELBs = append(region.Classic.ELBs, elb)
			elb.Classic = region.Classic
		}

		for _, azname := range elb.AvailabilityZoneNames {
			if az, exists := availability_zones[azname]; exists {
				elb.AvailabilityZones = append(elb.AvailabilityZones, az)
			}
		}

		for _, subnetid := range elb.SubnetIds {
			if subnet, exists := subnets[subnetid]; exists {
				elb.Subnets = append(elb.Subnets, subnet)
			}
		}

		for _, elbInst := range elb.ELBInstances {
			if elbInst.InstanceId != nil {
				if inst, exists := instances[*elbInst.InstanceId]; exists {
					inst.ELB = elb
					elb.Instances = append(elb.Instances, inst)
				}
			}
		}

		for _, sgname := range elb.SecurityGroupNames {
			for _, sg := range security_groups {
				if sg.GroupName == sgname {
					elb.SecurityGroups = append(elb.SecurityGroups, sg)
					sg.ELBs = append(sg.ELBs, elb)
					break
				}
			}
		}

		if len(elb.SourceSecurityGroupName) > 0 {
			if sg, exists := security_groups[elb.SourceSecurityGroupName]; exists {
				elb.SourceSecurityGroup = sg
				sg.ELBs = append(sg.ELBs, elb)
			}
		}

	}

	for _, group := range as_groups {
		region.AutoScalingGroups = append(region.AutoScalingGroups, group)
		for _, group_instance := range group.AutoScalingInstances {
			if group_instance.InstanceId != nil {
				if inst, exists := instances[*group_instance.InstanceId]; exists {
					group.Instances = append(group.Instances, inst)
					inst.AutoScalingGroup = group
					if !AvailabilityZoneInSlice(group.AvailabilityZones, inst.AvailabilityZone) {
						group.AvailabilityZones = append(group.AvailabilityZones, inst.AvailabilityZone)
					}
					if inst.ELB != nil && !AutoScalingGroupInSlice(inst.ELB.AutoScalingGroups, group) {
						inst.ELB.AutoScalingGroups = append(inst.ELB.AutoScalingGroups, group)
					}
					if inst.VPC != nil && !AutoScalingGroupInSlice(inst.VPC.AutoScalingGroups, group) {
						inst.VPC.AutoScalingGroups = append(inst.VPC.AutoScalingGroups, group)
					} else if inst.Classic != nil && !AutoScalingGroupInSlice(inst.Classic.AutoScalingGroups, group) {
						inst.Classic.AutoScalingGroups = append(inst.Classic.AutoScalingGroups, group)
					}
				}
			}
		}
	}

	for _, table := range route_tables {
		for _, assoc := range table.Associations {
			if assoc.SubnetId != nil {
				if subnet, exists := subnets[*assoc.SubnetId]; exists {
					subnet.RouteTables = append(subnet.RouteTables, table)
				}
			}
		}
		if vpc, exists := vpcs[table.VpcId]; exists {
			vpc.RouteTables = append(vpc.RouteTables, table)
		}
	}

	for _, acl := range acls {
		if vpc, exists := vpcs[acl.VpcId]; exists {
			vpc.ACLs = append(vpc.ACLs, acl)
		}
		for _, assoc := range acl.Associations {
			if assoc.SubnetId != nil {
				if subnet, exists := subnets[*assoc.SubnetId]; exists {
					subnet.ACLs = append(subnet.ACLs, acl)
				}
			}
		}
	}

	for _, vpce := range vpc_endpoints {
		if vpc, exists := vpcs[vpce.VpcId]; exists {
			vpce.VPC = vpc
			vpc.VPCEndpoints = append(vpc.VPCEndpoints, vpce)
		}
	}

	for _, vpcp := range vpc_peering_connections {
		if vpcp.RequesterVpcInfo != nil && vpcp.RequesterVpcInfo.VpcId != nil {
			if vpc, exists := vpcs[*vpcp.RequesterVpcInfo.VpcId]; exists {
				vpcp.RequesterVPC = vpc
				vpc.VPCPeeringConnections = append(vpc.VPCPeeringConnections, vpcp)
			}
		}
		if vpcp.AccepterVpcInfo != nil && vpcp.AccepterVpcInfo.VpcId != nil {
			if vpc, exists := vpcs[*vpcp.AccepterVpcInfo.VpcId]; exists {
				vpcp.AccepterVPC = vpc
				vpc.VPCPeeringConnections = append(vpc.VPCPeeringConnections, vpcp)
			}
		}
	}

	for _, ig := range internet_gateways {
		region.InternetGateways = append(region.InternetGateways, ig)
		if len(ig.Attachments) > 0 && ig.Attachments[0].VpcId != nil {
			if vpc, exists := vpcs[*ig.Attachments[0].VpcId]; exists {
				vpc.InternetGateway = ig
				ig.VPCs = append(ig.VPCs, vpc)
			}
		}
	}

	for _, cgw := range customer_gateways {
		region.CustomerGateways = append(region.CustomerGateways, cgw)
	}

	for _, vpg := range vp_gateways {
		region.VPGateways = append(region.VPGateways, vpg)
		for _, attachment := range vpg.VpcAttachments {
			if attachment.VpcId != nil {
				if vpc, exists := vpcs[*attachment.VpcId]; exists {
					vpc.VPGateways = append(vpc.VPGateways, vpg)
					vpg.VPCs = append(vpg.VPCs, vpc)
				}
			}
		}
	}

	for _, vpn := range vpn_connections {
		if cgw, exists := customer_gateways[vpn.CustomerGatewayId]; exists {
			vpn.CustomerGateway = cgw
		}
		if vpg, exists := vp_gateways[vpn.VpnGatewayId]; exists {
			vpn.VPGateway = vpg
			vpg.VPNConnections = append(vpg.VPNConnections, vpn)
			for _, attachment := range vpg.VpcAttachments {
				if attachment.VpcId != nil {
					if vpc, exists := vpcs[*attachment.VpcId]; exists {
						vpc.VPNConnections = append(vpc.VPNConnections, vpn)
						if vpn.CustomerGateway != nil {
							vpc.CustomerGateways = append(vpc.CustomerGateways, vpn.CustomerGateway)
						}
					}
				}
			}
		}
		region.VPNConnections = append(region.VPNConnections, vpn)
	}

	for _, subnet := range subnets {
		if az, exists := availability_zones[subnet.AvailabilityZoneName]; exists {
			az.Subnets = append(az.Subnets, subnet)
			subnet.AvailabilityZone = az
		}
		for _, table := range subnet.RouteTables {
			if subnet.VPC == nil {
				if vpc, exists := vpcs[table.VpcId]; exists {
					subnet.VPC = vpc
					vpc.Subnets = append(vpc.Subnets, subnet)
					if vpcaz, exists := vpc.azs[subnet.AvailabilityZone]; exists {
						vpcaz.Subnets = append(vpcaz.Subnets, subnet)
					} else {
						vpc.azs[subnet.AvailabilityZone] = &AvailabilityZone{
							Messages:   subnet.AvailabilityZone.Messages,
							RegionName: subnet.AvailabilityZone.RegionName,
							State:      subnet.AvailabilityZone.State,
							ZoneName:   subnet.AvailabilityZone.ZoneName,
							Name:       subnet.AvailabilityZone.Name,
							Subnets:    []*Subnet{subnet},
						}
					}
				}
			}
			for _, route := range table.Routes {
				switch {
				case route.GatewayId != nil:
					if ig, exists := internet_gateways[*route.GatewayId]; exists {
						subnet.InternetGateway = ig
					} else if vpce, exists := vpc_endpoints[*route.GatewayId]; exists {
						if !VPCEndpointInSlice(subnet.VPCEndpoints, vpce) {
							subnet.VPCEndpoints = append(subnet.VPCEndpoints, vpce)
							vpce.Subnets = append(vpce.Subnets, subnet)
						}
					} else if vpg, exists := vp_gateways[*route.GatewayId]; exists {
						if !SubnetInSlice(vpg.Subnets, subnet) {
							vpg.Subnets = append(vpg.Subnets, subnet)
						}
						for _, vpn := range vpn_connections {
							if vpn.VpnGatewayId == *route.GatewayId {
								subnet.VPNConnections = append(subnet.VPNConnections, vpn)
								break
							}
						}
					}
				case route.VpcPeeringConnectionId != nil:
					if vpcp, exists := vpc_peering_connections[*route.VpcPeeringConnectionId]; exists {
						subnet.VPCPeeringConnections = append(subnet.VPCPeeringConnections, vpcp)
						vpcp.Subnets = append(vpcp.Subnets, subnet)
					}
				case route.InstanceId != nil:
					// source/dest check is characteristically disabled on nat instances, but not others
					if inst, exists := instances[*route.InstanceId]; exists && !inst.SourceDestCheck {
						subnet.NATInstance = inst
					}
				case route.NatGatewayId != nil:
					if nat, exists := nat_gateways[*route.NatGatewayId]; exists {
						subnet.NATGateway = nat
						nat.Subnet = subnet
						subnet.VPC.NATGateways = append(subnet.VPC.NATGateways, nat)
					}
				}
			}
		}
	}

	for _, ami := range amis {
		region.AMIs = append(region.AMIs, ami)
	}
	for _, sg := range security_groups {
		region.SecurityGroups = append(region.SecurityGroups, sg)
	}

	for _, vpc := range vpcs {
		for _, inst := range vpc.Instances {
			if inst.AMI != nil && !AMIInSlice(vpc.AMIs, inst.AMI) {
				vpc.AMIs = append(vpc.AMIs, inst.AMI)
			}
			if inst.AvailabilityZone != nil {
				if az, exists := vpc.azs[inst.AvailabilityZone]; exists {
					az.Instances = append(az.Instances, inst)
				}
			}
			for _, sg := range inst.SecurityGroups {
				if !SecurityGroupInSlice(vpc.SecurityGroups, sg) {
					vpc.SecurityGroups = append(vpc.SecurityGroups, sg)
				}
				if !VPCInSlice(sg.VPCs, vpc) {
					sg.VPCs = append(sg.VPCs, vpc)
				}
			}
		}
		for _, dbinst := range vpc.DBInstances {
			if dbinst.AvailabilityZone != nil {
				if az, exists := vpc.azs[dbinst.AvailabilityZone]; exists {
					az.DBInstances = append(az.DBInstances, dbinst)
				} else {
					vpc.azs[dbinst.AvailabilityZone] = &AvailabilityZone{
						Messages:    dbinst.AvailabilityZone.Messages,
						RegionName:  dbinst.AvailabilityZone.RegionName,
						State:       dbinst.AvailabilityZone.State,
						ZoneName:    dbinst.AvailabilityZone.ZoneName,
						Name:        dbinst.AvailabilityZone.Name,
						DBInstances: []*DBInstance{dbinst},
					}
				}
			}
		}
		for _, az := range vpc.azs {
			vpc.AvailabilityZones = append(vpc.AvailabilityZones, az)
		}
		region.VPCs = append(region.VPCs, vpc)
		vpc.Region = region
	}

	func() {
		localazs := map[*AvailabilityZone]*AvailabilityZone{}

		for _, inst := range region.Classic.Instances {
			if inst.AMI != nil && !AMIInSlice(region.Classic.AMIs, inst.AMI) {
				region.Classic.AMIs = append(region.Classic.AMIs, inst.AMI)
			}
			if inst.AvailabilityZone != nil {
				if az, exists := localazs[inst.AvailabilityZone]; exists {
					az.Instances = append(az.Instances, inst)
				} else {
					localazs[inst.AvailabilityZone] = &AvailabilityZone{
						Messages:   inst.AvailabilityZone.Messages,
						RegionName: inst.AvailabilityZone.RegionName,
						State:      inst.AvailabilityZone.State,
						ZoneName:   inst.AvailabilityZone.ZoneName,
						Name:       inst.AvailabilityZone.Name,
						Instances:  []*Instance{inst},
					}
				}
			}
			for _, sg := range inst.SecurityGroups {
				if !SecurityGroupInSlice(region.Classic.SecurityGroups, sg) {
					region.Classic.SecurityGroups = append(region.Classic.SecurityGroups, sg)
				}
				sg.Classic = region.Classic
			}
		}
		for _, az := range localazs {
			region.Classic.AvailabilityZones = append(region.Classic.AvailabilityZones, az)
		}
		region.Classic.Region = region
	}()

	for _, inst := range db_instances {
		region.DBInstances = append(region.DBInstances, inst)
		inst.Region = region
		if az, exists := availability_zones[inst.AvailabilityZoneName]; exists {
			az.DBInstances = append(az.DBInstances, inst)
			inst.AvailabilityZone = az
		}
		if inst.DBSubnetGroup != nil && inst.DBSubnetGroup.VpcId != nil {
			if vpc, exists := vpcs[*inst.DBSubnetGroup.VpcId]; exists {
				vpc.DBInstances = append(vpc.DBInstances, inst)
				inst.VPC = vpc
			}
		} else {
			region.Classic.DBInstances = append(region.Classic.DBInstances, inst)
			inst.Classic = region.Classic
		}
		for _, group := range inst.VpcSecurityGroups {
			if group.VpcSecurityGroupId != nil {
				if sg, exists := security_groups[*group.VpcSecurityGroupId]; exists {
					sg.DBInstances = append(sg.DBInstances, inst)
					inst.SecurityGroups = append(inst.SecurityGroups)
				}
			}
		}
	}

	for _, vpc := range vpcs {
		for _, dbinst := range vpc.DBInstances {
			if az, exists := vpc.azs[dbinst.AvailabilityZone]; exists {
				az.DBInstances = append(az.DBInstances, dbinst)
			} else {
				vpc.azs[dbinst.AvailabilityZone] = &AvailabilityZone{
					Messages:    dbinst.AvailabilityZone.Messages,
					RegionName:  dbinst.AvailabilityZone.RegionName,
					State:       dbinst.AvailabilityZone.State,
					ZoneName:    dbinst.AvailabilityZone.ZoneName,
					Name:        dbinst.AvailabilityZone.Name,
					DBInstances: []*DBInstance{dbinst},
				}
			}
		}
	}

	for _, ecc := range ec_clusters {
		region.ElasticCacheClusters = append(region.ElasticCacheClusters, ecc)
		ecc.Region = region
		for _, node := range ecc.CacheNodes {
			if azname := node.CustomerAvailabilityZone; azname != nil {
				if az, exists := availability_zones[*azname]; exists {
					if !ElasticCacheClusterInSlice(az.ElasticCacheClusters, ecc) {
						az.ElasticCacheClusters = append(az.ElasticCacheClusters, ecc)
					}
					if !AvailabilityZoneInSlice(ecc.AvailabilityZones, az) {
						ecc.AvailabilityZones = append(ecc.AvailabilityZones, az)
					}
				}
			}
		}
		if len(ecc.VPCSecurityGroups) > 0 {
			for _, group := range ecc.VPCSecurityGroups {
				if group.SecurityGroupId != nil {
					if sg, exists := security_groups[*group.SecurityGroupId]; exists {
						sg.ElasticCacheClusters = append(sg.ElasticCacheClusters, ecc)
						ecc.SecurityGroups = append(ecc.SecurityGroups, sg)
						if vpc, exists := vpcs[sg.VpcId]; exists {
							vpc.ElasticCacheClusters = append(vpc.ElasticCacheClusters, ecc)
							ecc.VPC = vpc
							break
						}
					}
				}
			}
		} else {
			region.Classic.ElasticCacheClusters = append(region.Classic.ElasticCacheClusters, ecc)
			ecc.Classic = region.Classic
		}
		for _, group := range ecc.CacheSecurityGroups {
			if group.CacheSecurityGroupName != nil {
				if sg, exists := security_groups[*group.CacheSecurityGroupName]; exists {
					sg.ElasticCacheClusters = append(sg.ElasticCacheClusters, ecc)
					ecc.SecurityGroups = append(ecc.SecurityGroups, sg)
				}
			}
		}
	}

	for _, az := range availability_zones {
		region.AvailabilityZones = append(region.AvailabilityZones, az)
	}

	for _, lf := range lambda_functions {
		lf.Region = region
		region.LambdaFunctions = append(region.LambdaFunctions, lf)
		if lf.VpcConfig != nil {
			if lf.VpcConfig.VpcId != nil {
				if vpc, exists := vpcs[*lf.VpcConfig.VpcId]; exists {
					vpc.LambdaFunctions = append(vpc.LambdaFunctions, lf)
					lf.VPC = vpc
				}
			}
			for _, sgid := range lf.VpcConfig.SecurityGroupIds {
				if sgid != nil {
					if sg, exists := security_groups[*sgid]; exists {
						sg.LambdaFunctions = append(sg.LambdaFunctions, lf)
						lf.SecurityGroups = append(lf.SecurityGroups, sg)
					}
				}
			}
			for _, sid := range lf.VpcConfig.SubnetIds {
				if sid != nil {
					if subnet, exists := subnets[*sid]; exists {
						subnet.LambdaFunctions = append(subnet.LambdaFunctions, lf)
						lf.Subnets = append(lf.Subnets, subnet)
					}
				}
			}
		}
	}
	for _, eni := range enis {
		if eni.Attachment != nil && eni.Attachment.InstanceId != nil {
			if inst, exists := instances[aws.StringValue(eni.Attachment.InstanceId)]; exists {
				inst.ENIs = append(inst.ENIs, eni)
			}
		}
		if vpc, exists := vpcs[eni.VpcId]; exists {
			vpc.ENIs = append(vpc.ENIs, eni)
			eni.VPC = vpc
			if len(eni.AvailabilityZone) > 0 {
				for _, az := range vpc.AvailabilityZones {
					if az.ZoneName == eni.AvailabilityZone {
						az.ENIs = append(az.ENIs, eni)
						break
					}
				}
			}
		}
	}
	for _, queue := range sqs_queues {
		queue.Region = region
		region.SQSQueues = append(region.SQSQueues, queue)
	}
	for _, topic := range sns_topics {
		topic.Region = region
		region.SNSTopics = append(region.SNSTopics, topic)
	}
	for _, sub := range sns_subscribers {
		sub.Region = region
		region.SNSSubscriptions = append(region.SNSSubscriptions, sub)
		if topic, exists := sns_topics[sub.TopicArn]; exists {
			topic.Subscribers = append(topic.Subscribers, sub)
		}
	}
	for _, nat := range nat_gateways {
		nat.Region = region
		if vpc, exists := vpcs[nat.VpcId]; exists {
			nat.VPC = vpc
		}
	}
	for _, alarm := range cloudwatch_alarms {
		alarm.Region = region
		region.CloudWatchAlarms = append(region.CloudWatchAlarms, alarm)
		for _, d := range alarm.Dimensions {
			switch aws.StringValue(d.Name) {
			case "InstanceId":
				if inst, exists := instances[aws.StringValue(d.Value)]; exists {
					inst.CloudWatchAlarms = append(inst.CloudWatchAlarms, alarm)
				}
			case "DBInstanceIdentifier":
				if dbinst, exists := db_instances[aws.StringValue(d.Value)]; exists {
					dbinst.CloudWatchAlarms = append(dbinst.CloudWatchAlarms, alarm)
				}
			case "CacheClusterId":
				if ecc, exists := ec_clusters[aws.StringValue(d.Value)]; exists {
					ecc.CloudWatchAlarms = append(ecc.CloudWatchAlarms, alarm)
				}
			case "AutoScalingGroupName":
				if group, exists := as_groups[aws.StringValue(d.Value)]; exists {
					group.CloudWatchAlarms = append(group.CloudWatchAlarms, alarm)
				}
			case "LoadBalancerName":
				if elb, exists := elbs[aws.StringValue(d.Value)]; exists {
					elb.CloudWatchAlarms = append(elb.CloudWatchAlarms, alarm)
				}
			case "FunctionName":
				if lf, exists := lambda_functions[aws.StringValue(d.Value)]; exists {
					lf.CloudWatchAlarms = append(lf.CloudWatchAlarms, alarm)
				}
			case "MountPath", "Filesystem", "CacheNodeId":
				// ignore, caught by other dimension
			default:
				// fmt.Println("unknown dimension", d)
			}
		}
		for _, arn := range alarm.AlarmActions {
			switch {
			case strings.Contains(arn, ":sns:"):
				for _, sns := range sns_topics {
					if sns.TopicArn == arn {
						alarm.AlarmActionSNSs = append(alarm.AlarmActionSNSs, sns)
						break
					}
				}
			case strings.Contains(arn, ":autoscaling:"):
				for _, ag := range as_groups {
					if ag.AutoScalingGroupARN == arn {
						alarm.AlarmActionAutoScalingGroups = append(alarm.AlarmActionAutoScalingGroups, ag)
						break
					}
				}
			}
		}
		for _, arn := range alarm.InsufficientDataActions {
			switch {
			case strings.Contains(arn, ":sns:"):
				for _, sns := range sns_topics {
					if sns.TopicArn == arn {
						alarm.InsufficientDataActionSNSs = append(alarm.InsufficientDataActionSNSs, sns)
						break
					}
				}
			case strings.Contains(arn, ":autoscaling:"):
				for _, ag := range as_groups {
					if ag.AutoScalingGroupARN == arn {
						alarm.InsufficientDataActionAutoScalingGroups = append(alarm.InsufficientDataActionAutoScalingGroups, ag)
						break
					}
				}
			}
		}
		for _, arn := range alarm.OKActions {
			switch {
			case strings.Contains(arn, ":sns:"):
				for _, sns := range sns_topics {
					if sns.TopicArn == arn {
						alarm.OKActionSNSs = append(alarm.OKActionSNSs, sns)
						break
					}
				}
			case strings.Contains(arn, ":autoscaling:"):
				for _, ag := range as_groups {
					if ag.AutoScalingGroupARN == arn {
						alarm.OKActionAutoScalingGroups = append(alarm.OKActionAutoScalingGroups, ag)
						break
					}
				}
			}
		}
	}

	for _, lf := range lambda_functions {
		sort.Sort(SecurityGroupByNameAsc(lf.SecurityGroups))
		sort.Sort(SubnetByNameAsc(lf.Subnets))
	}
	for _, topic := range sns_topics {
		sort.Sort(SNSSubscriptionByNameAsc(topic.Subscribers))
	}
	for _, vpg := range vp_gateways {
		sort.Sort(SubnetByCIDRAsc(vpg.Subnets))
	}
	for _, vpcs := range vpc_peering_connections {
		sort.Sort(SubnetByCIDRAsc(vpcs.Subnets))
	}
	for _, vpce := range vpc_endpoints {
		sort.Sort(SubnetByCIDRAsc(vpce.Subnets))
	}
	for _, ami := range amis {
		sort.Sort(InstanceByNameAsc(ami.Instances))
	}
	for _, sg := range security_groups {
		sort.Sort(SecurityGroupIpPermissionsByPortAsc(sg.IpPermissions))
		sort.Sort(SecurityGroupIpPermissionsByPortAsc(sg.IpPermissionsEgress))
	}
	for _, ecc := range ec_clusters {
		sort.Sort(AvailabilityZoneByNameAsc(ecc.AvailabilityZones))
	}
	for _, az := range availability_zones {
		sort.Sort(DBInstanceByNameAsc(az.DBInstances))
		sort.Sort(ElasticCacheClusterByNameAsc(az.ElasticCacheClusters))
		sort.Sort(InstanceByNameAsc(az.Instances))
		sort.Sort(SubnetByCIDRAsc(az.Subnets))
	}
	for _, elb := range elbs {
		sort.Sort(AvailabilityZoneByNameAsc(elb.AvailabilityZones))
		sort.Sort(InstanceByNameAsc(elb.Instances))
	}
	for _, inst := range instances {
		sort.Sort(ENIByNameAsc(inst.ENIs))
		sort.Sort(SecurityGroupByNameAsc(inst.SecurityGroups))
	}
	for _, subnet := range subnets {
		sort.Sort(ACLByNameAsc(subnet.ACLs))
		sort.Sort(InstanceByNameAsc(subnet.Instances))
		sort.Sort(RouteTableByNameAsc(subnet.RouteTables))
		sort.Sort(VPCEndpointByNameAsc(subnet.VPCEndpoints))
		sort.Sort(VPCPeeringConnectionByNameAsc(subnet.VPCPeeringConnections))
		sort.Sort(VPNConnectionByNameAsc(subnet.VPNConnections))
	}

	// region
	sort.Sort(AMIByNameAsc(region.AMIs))
	sort.Sort(AutoScalingGroupByNameAsc(region.AutoScalingGroups))
	sort.Sort(AvailabilityZoneByNameAsc(region.AvailabilityZones))
	sort.Sort(CloudWatchAlarmByNameAsc(region.CloudWatchAlarms))
	sort.Sort(CustomerGatewayByNameAsc(region.CustomerGateways))
	sort.Sort(DBInstanceByNameAsc(region.DBInstances))
	sort.Sort(ElasticCacheClusterByNameAsc(region.ElasticCacheClusters))
	sort.Sort(ELBByNameAsc(region.ELBs))
	sort.Sort(InstanceByNameAsc(region.Instances))
	sort.Sort(InternetGatewayByNameAsc(region.InternetGateways))
	sort.Sort(LambdaFunctionsByNameAsc(region.LambdaFunctions))
	sort.Sort(SecurityGroupByNameAsc(region.SecurityGroups))
	sort.Sort(SNSSubscriptionByNameAsc(region.SNSSubscriptions))
	sort.Sort(SNSTopicByNameAsc(region.SNSTopics))
	sort.Sort(SQSQueueByNameAsc(region.SQSQueues))
	sort.Sort(VPCByNameAsc(region.VPCs))
	sort.Sort(VPGatewayByNameAsc(region.VPGateways))
	sort.Sort(VPNConnectionByNameAsc(region.VPNConnections))

	// classic
	sort.Sort(AMIByNameAsc(region.Classic.AMIs))
	sort.Sort(AutoScalingGroupByNameAsc(region.Classic.AutoScalingGroups))
	sort.Sort(AvailabilityZoneByNameAsc(region.Classic.AvailabilityZones))
	sort.Sort(DBInstanceByNameAsc(region.Classic.DBInstances))
	sort.Sort(ElasticCacheClusterByNameAsc(region.Classic.ElasticCacheClusters))
	sort.Sort(ELBByNameAsc(region.Classic.ELBs))
	sort.Sort(InstanceByNameAsc(region.Classic.Instances))
	sort.Sort(SecurityGroupByNameAsc(region.Classic.SecurityGroups))

	// vpcs
	for _, vpc := range region.VPCs {
		sort.Sort(ACLByNameAsc(vpc.ACLs))
		sort.Sort(AMIByNameAsc(vpc.AMIs))
		sort.Sort(AutoScalingGroupByNameAsc(vpc.AutoScalingGroups))
		sort.Sort(AvailabilityZoneByNameAsc(vpc.AvailabilityZones))
		sort.Sort(DBInstanceByNameAsc(vpc.DBInstances))
		sort.Sort(ElasticCacheClusterByNameAsc(vpc.ElasticCacheClusters))
		sort.Sort(ELBByNameAsc(vpc.ELBs))
		sort.Sort(ENIByNameAsc(vpc.ENIs))
		sort.Sort(InstanceByNameAsc(vpc.Instances))
		sort.Sort(LambdaFunctionsByNameAsc(vpc.LambdaFunctions))
		sort.Sort(NATGatewaysByNameAsc(vpc.NATGateways))
		sort.Sort(RouteTableByNameAsc(vpc.RouteTables))
		sort.Sort(SecurityGroupByNameAsc(vpc.SecurityGroups))
		sort.Sort(SubnetByCIDRAsc(vpc.Subnets))
		sort.Sort(VPCEndpointByNameAsc(vpc.VPCEndpoints))
		sort.Sort(VPCPeeringConnectionByNameAsc(vpc.VPCPeeringConnections))
		sort.Sort(VPNConnectionByNameAsc(vpc.VPNConnections))
		for _, az := range vpc.AvailabilityZones {
			sort.Sort(DBInstanceByNameAsc(az.DBInstances))
			sort.Sort(ENIByNameAsc(az.ENIs))
			sort.Sort(InstanceByNameAsc(az.Instances))
			sort.Sort(SubnetByCIDRAsc(az.Subnets))
		}
	}

	oldinsts := map[string]*Instance{}
	for _, oldinst := range prev_region.Instances {
		oldinsts[oldinst.InstanceId] = oldinst
	}
	for _, newinst := range region.Instances {
		if oldinst, exists := oldinsts[newinst.InstanceId]; exists {
			newinst.Unreachable = oldinst.Unreachable
			newinst.SysInfo = oldinst.SysInfo
			newinst.Stats = oldinst.Stats
		}
	}

	for _, v := range vpcs {
		region.Items[v.Id] = v
	}
	for _, v := range security_groups {
		region.Items[v.Id] = v
	}
	for _, v := range acls {
		region.Items[v.Id] = v
	}
	for _, v := range route_tables {
		region.Items[v.Id] = v
	}
	for _, v := range subnets {
		region.Items[v.Id] = v
	}
	for _, v := range elbs {
		region.Items[v.Id] = v
	}
	for _, v := range instances {
		region.Items[v.Id] = v
	}
	for _, v := range availability_zones {
		region.Items[v.Id] = v
	}
	for _, v := range internet_gateways {
		region.Items[v.Id] = v
	}
	for _, v := range customer_gateways {
		region.Items[v.Id] = v
	}
	for _, v := range vp_gateways {
		region.Items[v.Id] = v
	}
	for _, v := range vpn_connections {
		region.Items[v.Id] = v
	}
	for _, v := range vpc_endpoints {
		region.Items[v.Id] = v
	}
	for _, v := range vpc_peering_connections {
		region.Items[v.Id] = v
	}
	for _, v := range as_groups {
		region.Items[v.Id] = v
	}
	for _, v := range amis {
		region.Items[v.Id] = v
	}
	for _, v := range db_instances {
		region.Items[v.Id] = v
	}
	for _, v := range ec_clusters {
		region.Items[v.Id] = v
	}
	for _, v := range sqs_queues {
		region.Items[v.Id] = v
	}
	for _, v := range sns_topics {
		region.Items[v.Id] = v
	}
	for _, v := range cloudwatch_alarms {
		region.Items[v.Id] = v
	}
	for _, v := range lambda_functions {
		region.Items[v.Id] = v
	}
	for _, v := range enis {
		region.Items[v.Id] = v
	}
	for _, v := range nat_gateways {
		region.Items[v.Id] = v
	}

	fmt.Println("processing finished in", time.Since(start))
	start = time.Now()
	fmt.Println("collecting stats")

	var erraggregates [][][]chan error

	// erraggregates = append(erraggregates, region.RefreshInstances())
	erraggregates = append(erraggregates, region.RefreshElasticCacheClusters())
	erraggregates = append(erraggregates, region.RefreshELBs())
	erraggregates = append(erraggregates, region.RefreshLambdaFunctions())
	erraggregates = append(erraggregates, region.RefreshRDS())
	erraggregates = append(erraggregates, region.RefreshSNSTopics())
	erraggregates = append(erraggregates, region.RefreshSQSQueues())

	for _, a := range erraggregates {
		for _, b := range a {
			for _, c := range b {
				if err := <-c; err != nil {
					fmt.Println(err)
				}
			}
		}
	}

	fmt.Println("stats finished in", time.Since(start))
	tracker.Report()

	return nil

}

func (region *Region) RefreshLambdaFunctions() [][]chan error {

	var errs [][]chan error

	for _, lf := range region.LambdaFunctions {
		errs = append(errs, lf.Poll())
	}

	return errs

}

func (region *Region) RefreshSNSTopics() [][]chan error {

	var errs [][]chan error

	for _, t := range region.SNSTopics {
		errs = append(errs, t.Poll())
	}

	return errs

}

func (region *Region) RefreshSQSQueues() [][]chan error {

	var errs [][]chan error

	for _, q := range region.SQSQueues {
		errs = append(errs, q.Poll())
	}

	return errs

}

func (region *Region) RefreshElasticCacheClusters() [][]chan error {

	var errs [][]chan error

	for _, ecc := range region.ElasticCacheClusters {
		errs = append(errs, ecc.Poll())
	}

	return errs

}

func (region *Region) RefreshRDS() [][]chan error {

	var errs [][]chan error

	for _, db := range region.DBInstances {
		errs = append(errs, db.Poll())
	}

	return errs

}

func (region *Region) RefreshELBs() [][]chan error {

	var errs [][]chan error

	for _, elb := range region.ELBs {
		errs = append(errs, elb.Poll())
	}

	return errs

}

func (region *Region) RefreshInstances() [][]chan error {

	return nil

	var errs [][]chan error

	for _, inst := range region.Instances {
		errs = append(errs, inst.Poll())
	}

	fmt.Println("polling", len(errs), "instances")

	return errs

}

type (
	Graph struct {
		Vertices []*Vertex `json:"vertices"`
		Edges    []*Edge   `json:"edges"`
	}
	Vertex struct {
		Name string `json:"name"`
	}
	Edge struct {
		Source   int     `json:"source"`
		Target   int     `json:"target"`
		Directed bool    `json:"directed,omitempty"`
		Weight   float64 `json:"weight,omitempty"`
		Text     string  `json:"text,omitempty"`
	}
)

func (region *Region) BuildGraph() *Graph {

	var g Graph
	v := map[interface{}]int{}

	for _, cgw := range region.CustomerGateways {
		vtx := &Vertex{cgw.Name}
		v[cgw] = len(g.Vertices)
		g.Vertices = append(g.Vertices, vtx)
	}
	for _, vpg := range region.VPGateways {
		vtx := &Vertex{vpg.Name}
		v[vpg] = len(g.Vertices)
		g.Vertices = append(g.Vertices, vtx)
	}
	for _, vpn := range region.VPNConnections {
		vtx := &Vertex{vpn.Name}
		v[vpn] = len(g.Vertices)
		g.Vertices = append(g.Vertices, vtx)
	}

	for _, vpn := range region.VPNConnections {
		g.Edges = append(g.Edges, &Edge{
			Source:   v[vpn.CustomerGateway],
			Target:   v[vpn],
			Directed: true,
		})
		g.Edges = append(g.Edges, &Edge{
			Source:   v[vpn.VPGateway],
			Target:   v[vpn],
			Directed: true,
		})
	}

	return &g

}

type (
	Row struct {
		CustomerGateway *CustomerGateway
		VPNConnection   *VPNConnection
		VPGateway       *VPGateway
		VPC             *VPC
		InternetGateway *InternetGateway
	}
	rowOrder []*Row
)

func (a rowOrder) Len() int      { return len(a) }
func (a rowOrder) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a rowOrder) Less(i, j int) bool {

	acgw, bcgw := a[i].CustomerGateway, a[j].CustomerGateway
	if acgw != bcgw {
		var aname, bname string
		if acgw != nil {
			aname = acgw.Name
		}
		if bcgw != nil {
			bname = bcgw.Name
		}
		return string_less_than(aname, bname)
	}

	avpn, bvpn := a[i].VPNConnection, a[j].VPNConnection
	if avpn != bvpn {
		var aname, bname string
		if avpn != nil {
			aname = avpn.Name
		}
		if bvpn != nil {
			bname = bvpn.Name
		}
		return string_less_than(aname, bname)
	}

	avpg, bvpg := a[i].VPGateway, a[j].VPGateway
	if avpg != bvpg {
		var aname, bname string
		if avpg != nil {
			aname = avpg.Name
		}
		if bvpg != nil {
			bname = bvpg.Name
		}
		return string_less_than(aname, bname)
	}

	avpc, bvpc := a[i].VPC, a[j].VPC
	if avpc != bvpc {
		var aname, bname string
		if avpc != nil {
			aname = avpc.Name
		}
		if bvpc != nil {
			bname = bvpc.Name
		}
		return string_less_than(aname, bname)
	}

	aigw, bigw := a[i].InternetGateway, a[j].InternetGateway
	if aigw != bigw {
		var aname, bname string
		if aigw != nil {
			aname = aigw.Name
		}
		if bigw != nil {
			bname = bigw.Name
		}
		return string_less_than(aname, bname)
	}

	return false

}

func (region *Region) BuildConnTable() []*Row {

	g := map[*CustomerGateway]map[*VPNConnection]map[*VPGateway]map[*VPC]map[*InternetGateway]struct{}{}

	for _, cgw := range region.CustomerGateways {
		g[cgw] = map[*VPNConnection]map[*VPGateway]map[*VPC]map[*InternetGateway]struct{}{}
	}

	for _, vpn := range region.VPNConnections {
		if _, exists := g[vpn.CustomerGateway]; !exists {
			g[vpn.CustomerGateway] = map[*VPNConnection]map[*VPGateway]map[*VPC]map[*InternetGateway]struct{}{}
		}
		g[vpn.CustomerGateway][vpn] = map[*VPGateway]map[*VPC]map[*InternetGateway]struct{}{
			vpn.VPGateway: map[*VPC]map[*InternetGateway]struct{}{},
		}
	}

	for _, vpg := range region.VPGateways {
		if len(vpg.VPNConnections) == 0 {
			if _, exists := g[nil]; !exists {
				g[nil] = map[*VPNConnection]map[*VPGateway]map[*VPC]map[*InternetGateway]struct{}{}
			}
			if _, exists := g[nil][nil]; !exists {
				g[nil][nil] = map[*VPGateway]map[*VPC]map[*InternetGateway]struct{}{}
			}
			g[nil][nil][vpg] = map[*VPC]map[*InternetGateway]struct{}{}
		}
	}

	for _, vpc := range region.VPCs {
		if len(vpc.VPGateways) == 0 {
			if _, exists := g[nil]; !exists {
				g[nil] = map[*VPNConnection]map[*VPGateway]map[*VPC]map[*InternetGateway]struct{}{}
			}
			if _, exists := g[nil][nil]; !exists {
				g[nil][nil] = map[*VPGateway]map[*VPC]map[*InternetGateway]struct{}{}
			}
			if _, exists := g[nil][nil][nil]; !exists {
				g[nil][nil][nil] = map[*VPC]map[*InternetGateway]struct{}{}
			}
			g[nil][nil][nil][vpc] = map[*InternetGateway]struct{}{vpc.InternetGateway: struct{}{}}
		} else {
			for _, vpg := range vpc.VPGateways {
				if len(vpg.VPNConnections) == 0 {
					g[nil][nil][vpg][vpc] = map[*InternetGateway]struct{}{
						vpc.InternetGateway: struct{}{},
					}
				} else {
					for _, vpn := range vpg.VPNConnections {
						g[vpn.CustomerGateway][vpn][vpg][vpc] = map[*InternetGateway]struct{}{
							vpc.InternetGateway: struct{}{},
						}
					}
				}
			}
		}
	}

	for _, igw := range region.InternetGateways {
		if len(igw.VPCs) == 0 {
			if _, exists := g[nil]; !exists {
				g[nil] = map[*VPNConnection]map[*VPGateway]map[*VPC]map[*InternetGateway]struct{}{}
			}
			if _, exists := g[nil][nil]; !exists {
				g[nil][nil] = map[*VPGateway]map[*VPC]map[*InternetGateway]struct{}{}
			}
			if _, exists := g[nil][nil][nil]; !exists {
				g[nil][nil][nil] = map[*VPC]map[*InternetGateway]struct{}{}
			}
			if _, exists := g[nil][nil][nil][nil]; !exists {
				g[nil][nil][nil][nil] = map[*InternetGateway]struct{}{}
			}
			g[nil][nil][nil][nil][igw] = struct{}{}
		}
	}

	var rows []*Row

	// map[*CustomerGateway]map[*VPNConnection]map[*VPGateway]map[*VPC]map[*InternetGateway]struct{}{}

	for cgw, vpns := range g {
		for vpn, vpgs := range vpns {
			for vpg, vpcs := range vpgs {
				for vpc, igws := range vpcs {
					for igw, _ := range igws {
						rows = append(rows, &Row{
							CustomerGateway: cgw,
							VPNConnection:   vpn,
							VPGateway:       vpg,
							VPC:             vpc,
							InternetGateway: igw,
						})
					}
				}
			}
		}
	}

	sort.Sort(rowOrder(rows))

	return rows

}
