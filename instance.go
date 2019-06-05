package window

import (
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/emptyinterface/window/pricing"
	"github.com/emptyinterface/window/sysinfo"
	"golang.org/x/crypto/ssh"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type (
	Instance struct {

		// The AMI launch index, which can be used to find this instance in the launch
		// group.
		AmiLaunchIndex int64

		// The architecture of the image.
		Architecture string

		// Any block device mapping entries for the instance.
		BlockDeviceMappings []*ec2.InstanceBlockDeviceMapping

		// The idempotency token you provided when you launched the instance.
		ClientToken string

		// Indicates whether the instance is optimized for EBS I/O. This optimization
		// provides dedicated throughput to Amazon EBS and an optimized configuration
		// stack to provide optimal I/O performance. This optimization isn't available
		// with all instance types. Additional usage charges apply when using an EBS
		// Optimized instance.
		EbsOptimized bool

		// The hypervisor type of the instance.
		Hypervisor string

		// The IAM instance profile associated with the instance.
		IamInstanceProfile *ec2.IamInstanceProfile

		// The ID of the AMI used to launch the instance.
		ImageId string

		// The ID of the instance.
		InstanceId string

		// Indicates whether this is a Spot Instance.
		InstanceLifecycle string

		// The instance type.
		InstanceType string

		// The kernel associated with this instance.
		KernelId string

		// The name of the key pair, if this instance was launched with an associated
		// key pair.
		KeyName string

		// The time the instance was launched.
		LaunchTime time.Time

		// The monitoring information for the instance.
		Monitoring *ec2.Monitoring

		// [EC2-VPC] One or more network interfaces for the instance.
		NetworkInterfaces []*ec2.InstanceNetworkInterface

		// The location where the instance launched.
		Placement *ec2.Placement

		// The value is Windows for Windows instances; otherwise blank.
		Platform string

		// The private DNS name assigned to the instance. This DNS name can only be
		// used inside the Amazon EC2 network. This name is not available until the
		// instance enters the running state.
		PrivateDnsName string

		// The private IP address assigned to the instance.
		PrivateIpAddress string

		// The product codes attached to this instance.
		ProductCodes []*ec2.ProductCode

		// The public DNS name assigned to the instance. This name is not available
		// until the instance enters the running state.
		PublicDnsName string

		// The public IP address assigned to the instance.
		PublicIpAddress string

		// The RAM disk associated with this instance.
		RamdiskId string

		// The root device name (for example, /dev/sda1 or /dev/xvda).
		RootDeviceName string

		// The root device type used by the AMI. The AMI can use an EBS volume or an
		// instance store volume.
		RootDeviceType string

		// One or more security groups for the instance.
		SecurityGroupNames []*ec2.GroupIdentifier

		// Specifies whether to enable an instance launched in a VPC to perform NAT.
		// This controls whether source/destination checking is enabled on the instance.
		// A value of true means checking is enabled, and false means checking is disabled.
		// The value must be false for the instance to perform NAT. For more information,
		// see NAT Instances (http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_NAT_Instance.html)
		// in the Amazon Virtual Private Cloud User Guide.
		SourceDestCheck bool

		// The ID of the Spot Instance request.
		SpotInstanceRequestId string

		// Specifies whether enhanced networking is enabled.
		SriovNetSupport string

		// The current state of the instance.
		InstanceState *ec2.InstanceState

		// The reason for the most recent state transition.
		StateReason *ec2.StateReason

		// The reason for the most recent state transition. This might be an empty string.
		StateTransitionReason string

		// The ID of the subnet in which the instance is running.
		SubnetId string

		// Any tags assigned to the instance.
		Tags []*ec2.Tag

		// The virtualization type of the instance.
		VirtualizationType string

		// The ID of the VPC in which the instance is running.
		VpcId string

		Name             string
		Id               string
		State            string
		Region           *Region
		VPC              *VPC
		Classic          *Classic
		AvailabilityZone *AvailabilityZone
		Subnet           *Subnet
		ELB              *ELB
		SecurityGroups   []*SecurityGroup
		AMI              *AMI
		AutoScalingGroup *AutoScalingGroup
		ENIs             []*ENI
		CloudWatchAlarms []*CloudWatchAlarm

		// true if server cannot be ssh polled by usual means
		Unreachable       bool
		UnreachableReason string
		SysInfo           *sysinfo.SystemInfoCollector
		Stats             *sysinfo.SystemInfoSummary
		sysInfo_me        sync.RWMutex
	}

	InstanceByNameAsc         []*Instance
	NetworkInterfaceByNameAsc []*ec2.InstanceNetworkInterface
)

const DialTimeout = 3 * time.Second

func (a InstanceByNameAsc) Len() int      { return len(a) }
func (a InstanceByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a InstanceByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}
func (a NetworkInterfaceByNameAsc) Len() int      { return len(a) }
func (a NetworkInterfaceByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a NetworkInterfaceByNameAsc) Less(i, j int) bool {
	return string_less_than(*a[i].NetworkInterfaceId, *a[j].NetworkInterfaceId)
}

func (inst *Instance) CPU() int {
	if inst.Stats != nil && inst.Stats.CPU.PercentInUse > 0 {
		return int(inst.Stats.CPU.PercentInUse * 100)
	}
	return 0
}
func (inst *Instance) CPUs() []int {
	var cpus []int
	if inst.Stats != nil {
		for _, cpu := range inst.Stats.CPU.CPUs {
			if cpu.PercentInUse < 0 {
				cpus = append(cpus, 0)
			} else {
				cpus = append(cpus, int(cpu.PercentInUse*100))
			}
		}
	}
	return cpus
}
func (inst *Instance) MemoryUser() int {
	if inst.Stats != nil {
		return int(inst.Stats.Memory.PercentUser * 100)
	}
	return 0
}
func (inst *Instance) MemorySystem() int {
	if inst.Stats != nil {
		return int(inst.Stats.Memory.PercentSystem * 100)
	}
	return 0
}
func (inst *Instance) Disk() int {
	if inst.Stats != nil {
		return int(inst.Stats.Disk.PercentInUse * 100)
	}
	return 0
}
func (inst *Instance) NetworkIn() int {
	if inst.Stats != nil {
		return int(inst.Stats.Network.BytesPerSecondIn)
	}
	return 0
}
func (inst *Instance) NetworkOut() int {
	if inst.Stats != nil {
		return int(inst.Stats.Network.BytesPerSecondOut)
	}
	return 0
}
func (inst *Instance) NetworkInString() string {
	if inst.Stats != nil {
		return humanBytes(inst.Stats.Network.BytesPerSecondIn, 0)
	}
	return "0B"
}
func (inst *Instance) NetworkOutString() string {
	if inst.Stats != nil {
		return humanBytes(inst.Stats.Network.BytesPerSecondOut, 0)
	}
	return "0B"
}
func (inst *Instance) NetworkIO() string {
	if inst.Stats != nil {
		return fmt.Sprintf("%s / %s",
			humanBytes(inst.Stats.Network.BytesPerSecondIn, 0),
			humanBytes(inst.Stats.Network.BytesPerSecondOut, 0),
		)
	}
	return ""
}

func (inst *Instance) NetworkInNormal() float64 {
	const ceiling float64 = (125 << 20) / 10
	if inst.Stats != nil {
		return (float64(inst.Stats.Network.BytesPerSecondIn) / ceiling) * 100
	}
	return 0
}

func (inst *Instance) NetworkOutNormal() float64 {
	const ceiling float64 = (125 << 20) / 10
	if inst.Stats != nil {
		return (float64(inst.Stats.Network.BytesPerSecondOut) / ceiling) * 100
	}
	return 0
}

func LoadInstances(input *ec2.DescribeInstancesInput) (map[string]*Instance, error) {

	if input == nil {
		input = &ec2.DescribeInstancesInput{MaxResults: aws.Int64(1000)}
	} else {
		input.MaxResults = aws.Int64(1000) // max results
	}

	instances := map[string]*Instance{}

	if err := EC2Client.DescribeInstancesPages(input, func(page *ec2.DescribeInstancesOutput, _ bool) bool {
		for _, res := range page.Reservations {
			for _, ec2inst := range res.Instances {
				instance := &Instance{
					AmiLaunchIndex:        aws.Int64Value(ec2inst.AmiLaunchIndex),
					Architecture:          aws.StringValue(ec2inst.Architecture),
					BlockDeviceMappings:   ec2inst.BlockDeviceMappings,
					ClientToken:           aws.StringValue(ec2inst.ClientToken),
					EbsOptimized:          aws.BoolValue(ec2inst.EbsOptimized),
					Hypervisor:            aws.StringValue(ec2inst.Hypervisor),
					IamInstanceProfile:    ec2inst.IamInstanceProfile,
					ImageId:               aws.StringValue(ec2inst.ImageId),
					InstanceId:            aws.StringValue(ec2inst.InstanceId),
					InstanceLifecycle:     aws.StringValue(ec2inst.InstanceLifecycle),
					InstanceType:          aws.StringValue(ec2inst.InstanceType),
					KernelId:              aws.StringValue(ec2inst.KernelId),
					KeyName:               aws.StringValue(ec2inst.KeyName),
					LaunchTime:            aws.TimeValue(ec2inst.LaunchTime),
					Monitoring:            ec2inst.Monitoring,
					NetworkInterfaces:     ec2inst.NetworkInterfaces,
					Placement:             ec2inst.Placement,
					Platform:              aws.StringValue(ec2inst.Platform),
					PrivateDnsName:        aws.StringValue(ec2inst.PrivateDnsName),
					PrivateIpAddress:      aws.StringValue(ec2inst.PrivateIpAddress),
					ProductCodes:          ec2inst.ProductCodes,
					PublicDnsName:         aws.StringValue(ec2inst.PublicDnsName),
					PublicIpAddress:       aws.StringValue(ec2inst.PublicIpAddress),
					RamdiskId:             aws.StringValue(ec2inst.RamdiskId),
					RootDeviceName:        aws.StringValue(ec2inst.RootDeviceName),
					RootDeviceType:        aws.StringValue(ec2inst.RootDeviceType),
					SecurityGroupNames:    ec2inst.SecurityGroups,
					SourceDestCheck:       aws.BoolValue(ec2inst.SourceDestCheck),
					SpotInstanceRequestId: aws.StringValue(ec2inst.SpotInstanceRequestId),
					SriovNetSupport:       aws.StringValue(ec2inst.SriovNetSupport),
					InstanceState:         ec2inst.State,
					StateReason:           ec2inst.StateReason,
					StateTransitionReason: aws.StringValue(ec2inst.StateTransitionReason),
					SubnetId:              aws.StringValue(ec2inst.SubnetId),
					Tags:                  ec2inst.Tags,
					VirtualizationType:    aws.StringValue(ec2inst.VirtualizationType),
					VpcId:                 aws.StringValue(ec2inst.VpcId),
					sysInfo_me:            sync.RWMutex{},
				}
				instance.Name = TagOrDefault(instance.Tags, "Name", instance.InstanceId)
				instance.Id = "inst:" + instance.InstanceId
				instance.State = aws.StringValue(instance.InstanceState.Name)
				instances[instance.InstanceId] = instance
			}
		}
		return true
	}); err != nil {
		return nil, err
	}

	return instances, nil

}

func (inst *Instance) priceKey() string {
	var tenancy string
	if inst.Placement != nil {
		if inst.Placement.Tenancy != nil {
			switch *inst.Placement.Tenancy {
			case "dedicated":
				tenancy = pricing.DedicatedTenancy
			case "default":
				tenancy = pricing.SharedTenancy
			case "host":
				tenancy = pricing.HostTenancy
			}
		}
	}
	var platform string
	if len(inst.Platform) > 0 {
		platform = inst.Platform
	} else if inst.AMI != nil {
		switch {
		case strings.Contains(inst.AMI.Platform, "Windows"):
			platform = pricing.WindowsPlatform
		case strings.Contains(inst.AMI.Platform, "RHEL"):
			platform = pricing.RHELPlatform
		case strings.Contains(inst.AMI.Platform, "SUSE"):
			platform = pricing.SUSEPlatform
		default:
			platform = pricing.LinuxPlatform
		}
	} else {
		platform = pricing.LinuxPlatform
	}
	key := fmt.Sprintf("%s:%s:%s:%s:%s",
		pricing.AmazonEC2OfferCode,
		pricing.OnDemandTermType,
		tenancy,
		inst.InstanceType,
		platform,
	)
	return key
}

func (inst *Instance) HourlyCost() float64 {
	if offer, exists := inst.Region.Prices[inst.priceKey()]; exists {
		return offer.PricePerUnit
	}
	return 0
}

func (inst *Instance) MonthlyCost() float64 {
	if offer, exists := inst.Region.Prices[inst.priceKey()]; exists {
		return offer.PricePerUnit * 24 * 30
	}
	return 0
}

func (inst *Instance) Inactive() bool {
	return false
}

func (inst *Instance) Poll() []chan error {

	var errs []chan error

	if inst.SysInfo != nil {
		errs = append(errs, inst.Region.Throttle.do(inst.Name+" POLL", func() error {
			err := inst.SysInfo.Poll()
			inst.sysInfo_me.Lock()
			defer inst.sysInfo_me.Unlock()
			inst.Stats = inst.SysInfo.GetSummary()
			return err
		}))
		return errs
	}

	switch {
	case inst.Unreachable:
		return nil
	case aws.StringValue(inst.InstanceState.Name) != ec2.InstanceStateNameRunning:
		inst.UnreachableReason = "Instance not running"
		return nil
	case len(inst.KeyName) == 0:
		inst.Unreachable = true
		inst.UnreachableReason = "Key not specified"
		return nil
	}

	hosts := hostsForPort(inst, 22)
	if len(hosts) == 0 {
		inst.UnreachableReason = "No ports open"
		return nil
	}

	data, err := ioutil.ReadFile(filepath.Join(inst.Region.sshKeyPath, inst.KeyName+".pem"))
	if err != nil {
		inst.UnreachableReason = "Key file not found"
		return nil
	}

	signer, err := ssh.ParsePrivateKey(data)
	if err != nil {
		inst.UnreachableReason = "Key file not valid"
		return nil
	}

	// default to this
	inst.Unreachable = true
	inst.UnreachableReason = "No valid user found"

	users := []string{"ubuntu", "centos", "ec2-user", "admin"}

	try := func(user, host string) error {

		config := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
		}

		conn, err := net.DialTimeout("tcp", host, DialTimeout)
		if err != nil {
			return err
		}
		defer conn.Close()

		c, chans, reqs, err := ssh.NewClientConn(conn, host, config)
		if err != nil {
			return err
		}
		client := ssh.NewClient(c, chans, reqs)
		defer client.Close()

		sess, err := client.NewSession()
		if err != nil {
			return err
		}
		defer sess.Close()

		inst.sysInfo_me.Lock()
		defer inst.sysInfo_me.Unlock()
		if inst.SysInfo == nil {
			inst.SysInfo = sysinfo.NewSystemInfoCollector(host, config, 2)
			err := inst.SysInfo.Poll()
			inst.Stats = inst.SysInfo.GetSummary()
			inst.Unreachable = false
			inst.UnreachableReason = ""
			return err
		}

		return nil

	}

	for _, user := range users {
		for _, host := range hosts {
			user, host := user, host
			errs = append(errs, inst.Region.Throttle.do(inst.Name+" POLL", func() error {
				if err := try(user, host); err != nil && !strings.Contains(err.Error(), "ssh: handshake failed:") {
					if neterr, ok := err.(net.Error); !ok || (!neterr.Timeout() && !neterr.Temporary()) {
						return err
					}
				}
				return nil
			}))
		}
	}

	return errs

}

func (inst *Instance) PortsInvolved() []int {
	return SecurityGroupSet(inst.SecurityGroups).PortsInvolved()
}

func hostsForPort(inst *Instance, port int64) []string {

	pt := strconv.FormatInt(port, 10)
	var hosts []string

	if len(inst.NetworkInterfaces) > 0 {
		sgs := map[string]*SecurityGroup{}
		for _, sg := range inst.SecurityGroups {
			sgs[sg.GroupId] = sg
		}
		for _, ni := range inst.NetworkInterfaces {
			for _, group := range ni.Groups {
				if sg, exists := sgs[*group.GroupId]; exists {
					for _, perm := range sg.IpPermissions {
						if perm.ToPort != nil && *perm.ToPort == port {
							for _, priv := range ni.PrivateIpAddresses {
								if priv.PrivateIpAddress != nil {
									hosts = append(hosts, *priv.PrivateIpAddress+":"+pt)
								}
								if priv.Association != nil && priv.Association.PublicIp != nil {
									hosts = append(hosts, *priv.Association.PublicIp+":"+pt)
								}
							}
							goto NEXT_NETWORK_INTERFACE
						}
					}
				}
			}
		NEXT_NETWORK_INTERFACE:
		}
	} else {
		for _, sg := range inst.SecurityGroups {
			for _, perm := range sg.IpPermissions {
				if perm.ToPort != nil && *perm.ToPort == port {
					if len(inst.PrivateIpAddress) > 0 {
						hosts = append(hosts, inst.PrivateIpAddress+":"+pt)
					}
					if len(inst.PublicIpAddress) > 0 {
						hosts = append(hosts, inst.PublicIpAddress+":"+pt)
					}
				}
			}
		}
	}

	return hosts

}
