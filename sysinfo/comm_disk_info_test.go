package sysinfo

import (
	"reflect"
	"testing"
)

func TestDiskInfo(t *testing.T) {

	const output = `Filesystem      1B-blocks       Used  Available Use% Mounted on
/dev/xvda1     8578400256 1794719744 6783680512  21% /
devtmpfs        504578048          0  504578048   0% /dev
tmpfs           520523776          0  520523776   0% /dev/shm
tmpfs           520523776   58806272  461717504  12% /run
tmpfs           520523776          0  520523776   0% /sys/fs/cgroup
` + diskDelimeter + `Filesystem      Inodes IUsed   IFree IUse% Mounted on
/dev/xvda1     8387584 51593 8335991    1% /
devtmpfs        123188   269  122919    1% /dev
tmpfs           127081     1  127080    1% /dev/shm
tmpfs           127081   279  126802    1% /run
tmpfs           127081    13  127068    1% /sys/fs/cgroup
`

	expected := DiskInfo{Disk{Filesystem: "/dev/xvda1", MountPoint: "/", BytesTotal: 0x1ff500000, BytesUsed: 0x6af94000, BytesAvailable: 0x19456c000, PercentInUse: 0.20921380332477693, InodesTotal: 0x7ffc00, InodesUsed: 0xc989, InodesAvailable: 0x7f3277, PercentInodesInUse: 0.0061511157444146015}}

	stat := NewStat()

	if err := (&stat.DiskInfo).Parse([]byte(output)); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(stat.DiskInfo, expected) {
		t.Error("parse mismatch")
		dumpDiff(expected, stat.DiskInfo)
	}

}
