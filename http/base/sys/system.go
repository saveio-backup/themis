package sys

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common/fdlimit"
	"github.com/saveio/themis/common/log"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

var pid int
var once sync.Once
var SysScore = uint32(1)
var MonitorEnable = true

func InitOnce() {
	once.Do(func() {
		pid = os.Getpid()
		go Monitor()
	})
}

func Monitor() {
	var score float64
	for {
		if !MonitorEnable {
			return
		}
		time.Sleep(5 * time.Second)

		fdPercent, err := GetFdUsedPercent(int32(pid))
		if err != nil {
			log.Error("GetFdUsedPercent error: %s", err.Error())
			continue
		}
		cpuPercent, err := GetCpuUsedPercent()
		if err != nil {
			log.Error("GetCpuUsedPercent error: %s", err.Error())
			continue
		}
		memPercent, err := GetMemUsedPercent()
		if err != nil {
			log.Error("GetMemUsedPercent error: %s", err.Error())
			continue
		}
		score = 100 - (fdPercent+cpuPercent+memPercent)/3
		SysScore = uint32(score / 1)
	}
}

type SysUsedPercent struct {
	FdPercent  float64
	CPUPercent float64
	MemPercent float64
}

func GetSysUsedPercent() SysUsedPercent {
	lpid := os.Getpid()
	fdPercent, err := GetFdUsedPercent(int32(lpid))
	if err != nil {
		log.Error("GetFdUsedPercent error: %s", err.Error())
		return SysUsedPercent{}
	}
	cpuPercent, err := GetCpuUsedPercent()
	if err != nil {
		log.Error("GetCpuUsedPercent error: %s", err.Error())
		return SysUsedPercent{}
	}
	memPercent, err := GetMemUsedPercent()
	if err != nil {
		log.Error("GetMemUsedPercent error: %s", err.Error())
		return SysUsedPercent{}
	}
	return SysUsedPercent{
		FdPercent:  fdPercent,
		CPUPercent: cpuPercent,
		MemPercent: memPercent,
	}
}

func GetHostInfo() {
	hInfo, _ := host.Info()
	log.Infof("host info:%v uptime:%v boottime:%v", hInfo, hInfo.Uptime, hInfo.BootTime)
}

func GetCpuInfo() {
	cpuInfos, err := cpu.Info()
	if err != nil {
		log.Infof("Get cpu info failed, err:%v", err)
	}
	for _, ci := range cpuInfos {
		log.Infof("cpuInfo: %v", ci)
	}
}

func GetCpuLoad() {
	info, _ := load.Avg()
	log.Infof("%v", info)
}

func GetMemInfo() {
	memInfo, _ := mem.VirtualMemory()
	log.Infof("mem info:%v", memInfo)
}

func GetFdInfo() {
	maxSet, err := fdlimit.Maximum()
	if err != nil {
		log.Errorf("failed to get maximum open files: %v", err)
		return
	}

	curSet, err := fdlimit.Current()
	if err != nil {
		log.Errorf("failed to get current open files: %v", err)
		return
	}
	log.Infof("file descriptor maxSet: %d curSet: %d", maxSet, curSet)
}

func GetProcLinkCount(pid int32, protocol string) (int, error) {
	linkCount := 0
	states, err := net.ConnectionsPid(protocol, pid)
	if err != nil {
		return 0, fmt.Errorf("GetProcLinkCount error: %s", err.Error())
	}
	linkCount = len(states)
	log.Infof("GetProcLinkCount:%v ", linkCount)
	return linkCount, nil
}

func GetCpuUsedPercent() (float64, error) {
	percent, err := cpu.Percent(time.Second, false)
	log.Infof("cpu percent:%v", percent[0])
	return percent[0], err
}

func GetMemUsedPercent() (float64, error) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}
	log.Infof("mem percent:%v", memInfo.UsedPercent)
	return memInfo.UsedPercent, nil
}

func GetDiskInfo() {
	parts, err := disk.Partitions(true)
	if err != nil {
		log.Errorf("get Partitions failed, err:%v", err)
		return
	}
	for _, part := range parts {
		log.Infof("part:%v", part.String())
		diskInfo, err := disk.Usage(part.Mountpoint)
		if err == nil {
			log.Infof("disk info:used:%v free:%v", diskInfo.UsedPercent, diskInfo.Free)
		}

	}

	ioStat, _ := disk.IOCounters()
	for k, v := range ioStat {
		log.Infof("ioStat %v:%v", k, v)
	}
}

func GetNetInfo() {
	info, _ := net.IOCounters(true)
	for index, v := range info {
		log.Infof("%v:%v send:%v recv:%v", index, v, v.BytesSent, v.BytesRecv)
	}
}

func GetFdUsedPercent(pid int32) (float64, error) {
	curSet, err := fdlimit.Current()
	if err != nil {
		log.Errorf("failed to get current open files: %v", err)
		return 0, fmt.Errorf("failed to get current open files: %v", err)
	}

	fsUsedCount, err := GetFdUsedInfo(pid)
	if err != nil {
		log.Errorf("failed to get used fd info: %v", err)
		return 0, fmt.Errorf("failed to get used fd info: %v", err)
	}

	usedPercent := (float64(fsUsedCount) / float64(curSet)) * 100
	log.Infof("GetFdUsedPercent usedPercent :%v", usedPercent)
	return usedPercent, nil
}

func GetFdUsedInfo(pid int32) (uint64, error) {
	switch runtime.GOOS {
	case "darwin":
	case "linux":
	case "windows":
		return 80, nil
	}

	command := fmt.Sprintf("lsof -p %d | awk '{print $2}'| sort |uniq -c | grep %d | awk '{print $1}'", pid, pid)
	cmd := exec.Command("/bin/bash", "-c", command)

	stdout, _ := cmd.StdoutPipe()
	defer stdout.Close()
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("cmd.Start: %s", err.Error())
	}

	result, err := ioutil.ReadAll(stdout)
	if err != nil {
		return 0, fmt.Errorf("read stdout error: %s", err.Error())
	}
	resData := string(result)

	if err := cmd.Wait(); err != nil {
		if ex, ok := err.(*exec.ExitError); ok {
			res := ex.Sys().(syscall.WaitStatus).ExitStatus()
			if res != 0 {
				log.Errorf("GetFdUsedInfo cmd exit return: %d", res)
			}
		}
	}

	usedFileDescStr := strings.TrimSpace(resData)
	if len(usedFileDescStr) == 0 {
		return 0, nil
	}
	usedFileDesc, err := strconv.ParseUint(usedFileDescStr, 10, 64)
	if err != nil {
		err = fmt.Errorf("GetFdUsedInfo parseUint (%s) error: %s", usedFileDescStr, err.Error())
		log.Errorf("GetFdUsedInfo error: %s", err.Error())
	} else {
		log.Infof("GetFdUsedInfo, used: %d", usedFileDesc)
	}
	return usedFileDesc, err
}

//func main() {
//	var pid int
//	fmt.Println("Operation system: ", runtime.GOOS)
//
//	currProcPid := os.Getpid()
//	flag.IntVar(&pid, "pid", currProcPid, "example\n\t./main -pid=5")
//	flag.Parse()
//
//	//GetHostInfo()
//	//GetCpuInfo()
//	//GetMemInfo()
//	//GetFdInfo()
//	//GetDiskInfo()
//	//GetNetInfo()
//
//	GetFdUsedPercent(int32(pid))
//	GetProcLinkCount(int32(pid), "tcp")
//	GetCpuUsedPercent()
//	GetMemUsedPercent()
//}
