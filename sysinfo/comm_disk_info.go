package sysinfo

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"
)

type (
	Disk struct {
		Filesystem         string
		MountPoint         string
		BytesTotal         uint64
		BytesUsed          uint64
		BytesAvailable     uint64
		PercentInUse       float64
		InodesTotal        uint64
		InodesUsed         uint64
		InodesAvailable    uint64
		PercentInodesInUse float64
	}

	DiskInfo []Disk
)

const (
	diskDelimeter   = `===aRVZeDergP===`
	DiskInfoCommand = `df -B1 && echo -n ` + diskDelimeter + ` && df -i`
)

func (_ *DiskInfo) Command() string {
	return DiskInfoCommand
}

func (di *DiskInfo) Parse(b []byte) error {

	// output <df -B1 output> <delimiter> <df -i output>

	parts := bytes.Split(b, []byte(diskDelimeter))
	if len(parts) > 0 {
		if err := parseDFB1Output(di, parts[0]); err != nil {
			return err
		}
	}
	if len(parts) > 1 {
		if err := parseDFIOutput(di, parts[1]); err != nil {
			return err
		}
	}

	return nil

}

func parseDFB1Output(di *DiskInfo, b []byte) error {

	// output:
	// Filesystem      1B-blocks       Used  Available Use% Mounted on
	// /dev/xvda1     8578400256 1794719744 6783680512  21% /
	// devtmpfs        504578048          0  504578048   0% /dev
	// tmpfs           520523776          0  520523776   0% /dev/shm
	// tmpfs           520523776   58806272  461717504  12% /run
	// tmpfs           520523776          0  520523776   0% /sys/fs/cgroup

	sc := bufio.NewScanner(bytes.NewReader(b))
	if !sc.Scan() || sc.Err() != nil {
		return sc.Err()
	}

	headers := strings.Fields(sc.Text())

	for sc.Scan() {

		values := strings.Fields(sc.Text())

		if !strings.HasPrefix(values[0], "/dev/") {
			continue
		}

		var disk Disk

		for i, val := range values {
			switch headers[i] {
			case "Filesystem":
				disk.Filesystem = val
			case "Mounted":
				disk.MountPoint = val
			case "1B-blocks":
				disk.BytesTotal, _ = strconv.ParseUint(val, 10, 64)
			case "Used":
				disk.BytesUsed, _ = strconv.ParseUint(val, 10, 64)
			case "Available":
				disk.BytesAvailable, _ = strconv.ParseUint(val, 10, 64)
			}
		}

		disk.PercentInUse = float64(disk.BytesUsed) / float64(disk.BytesTotal)

		*di = append(*di, disk)

	}

	return nil

}

func parseDFIOutput(di *DiskInfo, b []byte) error {

	// output:
	// Filesystem      Inodes IUsed   IFree IUse% Mounted on
	// /dev/xvda1     8387584 51593 8335991    1% /
	// devtmpfs        123188   269  122919    1% /dev
	// tmpfs           127081     1  127080    1% /dev/shm
	// tmpfs           127081   279  126802    1% /run
	// tmpfs           127081    13  127068    1% /sys/fs/cgroup

	sc := bufio.NewScanner(bytes.NewReader(b))
	if !sc.Scan() || sc.Err() != nil {
		return sc.Err()
	}

	headers := strings.Fields(sc.Text())

	for sc.Scan() {

		values := strings.Fields(sc.Text())

		if !strings.HasPrefix(values[0], "/dev/") {
			continue
		}

		// search for existing disk to populate values for
		// use i to modify the value in-place
		for i, _ := range *di {
			disk := &(*di)[i] // grab a pointer to the disk struct
			if disk.Filesystem == values[0] {
				for i, val := range values {
					switch headers[i] {
					case "Inodes":
						disk.InodesTotal, _ = strconv.ParseUint(val, 10, 64)
					case "IUsed":
						disk.InodesUsed, _ = strconv.ParseUint(val, 10, 64)
					case "IFree":
						disk.InodesAvailable, _ = strconv.ParseUint(val, 10, 64)
					}
				}
				disk.PercentInodesInUse = float64(disk.InodesUsed) / float64(disk.InodesTotal)
				break
			}
		}

	}

	return nil

}
