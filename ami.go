package window

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type (
	AMI struct {
		// The architecture of the image.
		Architecture string

		// Any block device mapping entries.
		BlockDeviceMappings []*ec2.BlockDeviceMapping

		// The date and time the image was created.
		CreationDate time.Time

		// The description of the AMI that was provided during image creation.
		Description string

		// The hypervisor type of the image.
		Hypervisor string

		// The ID of the AMI.
		ImageId string

		// The location of the AMI.
		ImageLocation string

		// The AWS account alias (for example, amazon, self) or the AWS account ID of
		// the AMI owner.
		ImageOwnerAlias string

		// The type of image.
		ImageType string

		// The kernel associated with the image, if any. Only applicable for machine
		// images.
		KernelId string

		// The name of the AMI that was provided during image creation.
		Name string

		// The AWS account ID of the image owner.
		OwnerId string

		// The value is Windows for Windows AMIs; otherwise blank.
		Platform string

		// Any product codes associated with the AMI.
		ProductCodes []*ec2.ProductCode

		// Indicates whether the image has public launch permissions. The value is true
		// if this image has public launch permissions or false if it has only implicit
		// and explicit launch permissions.
		Public bool

		// The RAM disk associated with the image, if any. Only applicable for machine
		// images.
		RamdiskId string

		// The device name of the root device (for example, /dev/sda1 or /dev/xvda).
		RootDeviceName string

		// The type of root device used by the AMI. The AMI can use an EBS volume or
		// an instance store volume.
		RootDeviceType string

		// Specifies whether enhanced networking is enabled.
		SriovNetSupport string

		// The current state of the AMI. If the state is available, the image is successfully
		// registered and can be used to launch an instance.
		State string

		// The reason for the state change.
		StateReason *ec2.StateReason

		// Any tags assigned to the image.
		Tags []*ec2.Tag

		// The type of virtualization of the AMI.
		VirtualizationType string

		Id        string
		Instances []*Instance
	}

	AMIByNameAsc []*AMI
)

func (a AMIByNameAsc) Len() int      { return len(a) }
func (a AMIByNameAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a AMIByNameAsc) Less(i, j int) bool {
	return string_less_than(a[i].Name, a[j].Name)
}

func LoadAMIs(input *ec2.DescribeImagesInput) (map[string]*AMI, error) {

	resp, err := EC2Client.DescribeImages(input)
	if err != nil {
		return nil, err
	}

	images := make(map[string]*AMI, len(resp.Images))

	for _, img := range resp.Images {
		ami := &AMI{
			Architecture:        aws.StringValue(img.Architecture),
			BlockDeviceMappings: img.BlockDeviceMappings,
			Description:         aws.StringValue(img.Description),
			Hypervisor:          aws.StringValue(img.Hypervisor),
			ImageId:             aws.StringValue(img.ImageId),
			ImageLocation:       aws.StringValue(img.ImageLocation),
			ImageOwnerAlias:     aws.StringValue(img.ImageOwnerAlias),
			ImageType:           aws.StringValue(img.ImageType),
			KernelId:            aws.StringValue(img.KernelId),
			Name:                aws.StringValue(img.Name),
			OwnerId:             aws.StringValue(img.OwnerId),
			Platform:            aws.StringValue(img.Platform),
			ProductCodes:        img.ProductCodes,
			Public:              aws.BoolValue(img.Public),
			RamdiskId:           aws.StringValue(img.RamdiskId),
			RootDeviceName:      aws.StringValue(img.RootDeviceName),
			RootDeviceType:      aws.StringValue(img.RootDeviceType),
			SriovNetSupport:     aws.StringValue(img.SriovNetSupport),
			State:               aws.StringValue(img.State),
			StateReason:         img.StateReason,
			Tags:                img.Tags,
			VirtualizationType:  aws.StringValue(img.VirtualizationType),
		}
		ami.CreationDate, _ = time.Parse(time.RFC3339Nano, aws.StringValue(img.CreationDate))
		ami.Id = "ami:" + ami.ImageId
		images[ami.ImageId] = ami
	}

	return images, nil

}

func (ami *AMI) Inactive() bool {
	return len(ami.Instances) == 0
}
