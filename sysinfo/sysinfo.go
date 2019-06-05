package sysinfo

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type (
	Metric interface {
		Command() string
		Parse([]byte) error
	}

	SystemInfoCollector struct {
		Host   string
		Config *ssh.ClientConfig
		Stats  *StatSeries
	}

	SystemInfoSummary struct {
		Timestamp time.Time
		Duration  time.Duration
		Memory    struct {
			Total         uint64
			System        uint64
			User          uint64
			PercentSystem float64
			PercentUser   float64
		}
		Swap struct {
			Total        uint64
			Free         uint64
			PercentInUse float64
		}
		Disk Disk
		CPU  struct {
			CPUSummary
			ContextSwitches  uint64
			Processes        uint64
			ProcessesRunning uint64
			ProcessesBlocked uint64

			CPUs []CPUSummary
		}
		Network struct {
			BytesIn           uint64
			BytesOut          uint64
			BytesPerSecondIn  uint64
			BytesPerSecondOut uint64
		}
	}

	CPUSummary struct {
		Id               string
		PercentInUse     float64
		PercentUser      float64
		PercentNice      float64
		PercentSystem    float64
		PercentIdle      float64
		PercentIOWait    float64
		PercentIRQ       float64
		PercentSoftIRQ   float64
		PercentSteal     float64
		PercentGuest     float64
		PercentGuestNice float64
	}
)

const (
	RemoteBashCommand = `/bin/gzip -d | /bin/bash | /bin/gzip`
	commandDelimiter  = `===Jj52dgpmaF===`
	DialTimeout       = 2 * time.Second
)

var (
	// https://golang.org/src/compress/gzip/gunzip.go?s=#L177
	gzipHeaderByteSequence = []byte{0x1f, 0x8b, 0x08}
)

func NewSystemInfoCollector(host string, config *ssh.ClientConfig, entries int) *SystemInfoCollector {
	return &SystemInfoCollector{
		Host:   host,
		Config: config,
		Stats:  NewStatSeries(entries),
	}
}

func (si *SystemInfoCollector) Poll() error {

	stat := NewStat()

	if err := si.Execute(
		&stat.UpTime,
		&stat.CPUInfo,
		&stat.MemInfo,
		&stat.NetStat,
		&stat.LoadAvg,
		&stat.DiskInfo,
	); err != nil {
		return err
	}

	stat.Timestamp = time.Now()

	si.Stats.Add(stat)

	return nil

}

func (si *SystemInfoCollector) Execute(metrics ...Metric) error {

	conn, err := net.DialTimeout("tcp", si.Host, DialTimeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	c, chans, reqs, err := ssh.NewClientConn(conn, si.Host, si.Config)
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

	var (
		stdin  bytes.Buffer
		stdout bytes.Buffer
	)
	gzw := gzip.NewWriter(&stdin)
	gzw.Write(build_cmd(metrics))
	gzw.Close()

	sess.Stdin = &stdin
	sess.Stdout = &stdout

	if err := sess.Run(RemoteBashCommand); err != nil {
		return err
	}

	// look for gzip header and start from there so we can avoid
	// motd output and other shell noise
	headerStart := bytes.Index(stdout.Bytes(), gzipHeaderByteSequence)
	if headerStart < 0 {
		return gzip.ErrHeader
	}

	r, err := gzip.NewReader(bytes.NewReader(stdout.Bytes()[headerStart:]))
	if err != nil {
		return err
	}
	defer r.Close()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	stdouts := bytes.Split(data, []byte(commandDelimiter))

	for i, stdout := range stdouts {
		if err := metrics[i].Parse(stdout); err != nil {
			return err
		}
	}

	return nil

}

func build_cmd(metrics []Metric) (cmd []byte) {

	// glue commands together with delimiter
	const glue = " && echo -n " + commandDelimiter + " && "

	for i, m := range metrics {
		cmd = append(cmd, m.Command()...)
		if i < len(metrics)-1 {
			cmd = append(cmd, glue...)
		}
	}

	return

}

func (si *SystemInfoCollector) GetSummary() *SystemInfoSummary {

	var current, prev *Stat

	switch stats := si.Stats.Last(2); {
	case len(stats) == 2:
		prev, current = stats[0], stats[1]
	case len(stats) == 1:
		current = stats[0]
	default:
		return nil
	}

	s := &SystemInfoSummary{
		Timestamp: current.Timestamp,
	}

	// calculate memory use
	s.Memory.Total = current.MemInfo.MemTotal
	s.Memory.User = current.MemInfo.MemTotal - current.MemInfo.MemFree - current.MemInfo.Buffers - current.MemInfo.Cached
	s.Memory.System = current.MemInfo.MemTotal - current.MemInfo.MemFree - s.Memory.User
	s.Memory.PercentSystem = float64(s.Memory.System) / float64(s.Memory.Total)
	s.Memory.PercentUser = float64(s.Memory.User) / float64(s.Memory.Total)

	// calculate swap use
	s.Swap.Total = current.MemInfo.SwapTotal
	s.Swap.Free = current.MemInfo.SwapFree - current.MemInfo.SwapCached // this
	s.Swap.PercentInUse = 1 - (float64(s.Swap.Free) / float64(s.Swap.Total))

	// look for most in-use disk (inode or bytes)
	for _, disk := range current.DiskInfo {
		if disk.PercentInUse > s.Disk.PercentInUse || disk.PercentInodesInUse > s.Disk.PercentInodesInUse {
			s.Disk = disk
		}
	}

	// if we have something to do comparisons against, continue
	if prev != nil {
		s.Duration = current.UpTime.Total - prev.UpTime.Total

		// store last cpu stats
		s.CPU.ContextSwitches = current.CPUInfo.ContextSwitches - prev.CPUInfo.ContextSwitches
		s.CPU.Processes = current.CPUInfo.Processes - prev.CPUInfo.Processes
		s.CPU.ProcessesRunning = current.CPUInfo.ProcessesRunning
		s.CPU.ProcessesBlocked = current.CPUInfo.ProcessesBlocked

		// calculate aggregate stats
		calculate_cpu_stats(&s.CPU.CPUSummary, &prev.CPUInfo.CPUTotal, &current.CPUInfo.CPUTotal)

		// calculate stats per cpu
		for i := 0; i < len(prev.CPUInfo.CPUs); i++ {
			var cpu CPUSummary
			calculate_cpu_stats(&cpu, &prev.CPUInfo.CPUs[i], &current.CPUInfo.CPUs[i])
			s.CPU.CPUs = append(s.CPU.CPUs, cpu)
		}

		// calculate network
		seconds := s.Duration.Seconds()
		s.Network.BytesIn = current.NetStat.InOctets - prev.NetStat.InOctets
		s.Network.BytesOut = current.NetStat.OutOctets - prev.NetStat.OutOctets
		s.Network.BytesPerSecondIn = uint64(float64(s.Network.BytesIn) / seconds)
		s.Network.BytesPerSecondOut = uint64(float64(s.Network.BytesOut) / seconds)
	}

	return s

}

func calculate_cpu_stats(summary *CPUSummary, prev, current *CPU) {

	var (
		user      = float64(current.User - prev.User)
		nice      = float64(current.Nice - prev.Nice)
		system    = float64(current.System - prev.System)
		idle      = float64(current.Idle - prev.Idle)
		iowait    = float64(current.IOWait - prev.IOWait)
		irq       = float64(current.IRQ - prev.IRQ)
		softirq   = float64(current.SoftIRQ - prev.SoftIRQ)
		steal     = float64(current.Steal - prev.Steal)
		guest     = float64(current.Guest - prev.Guest)
		guestnice = float64(current.GuestNice - prev.GuestNice)
		total     = (user + nice + system + idle + iowait + irq + softirq + steal + guest + guestnice)
	)

	summary.Id = current.Id
	summary.PercentUser = user / total
	summary.PercentNice = nice / total
	summary.PercentSystem = system / total
	summary.PercentIdle = idle / total
	summary.PercentIOWait = iowait / total
	summary.PercentIRQ = irq / total
	summary.PercentSoftIRQ = softirq / total
	summary.PercentSteal = steal / total
	summary.PercentGuest = guest / total
	summary.PercentGuestNice = guestnice / total
	summary.PercentInUse = (total - idle - iowait) / total

}
