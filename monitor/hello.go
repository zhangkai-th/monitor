package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/shirou/gopsutil/disk"
)

func hello() {
	parts, err := disk.Partitions(true)
	if err != nil {
		fmt.Errorf("get Partitions failed,Error:%v", err)
	}
	label := fmt.Sprintf("\x1B[4;32;40mDevice\t\tFree\t\ttotal\t\tPercent(%%)\t\t\x1B[0m\n")
	io.WriteString(os.Stdout, label)
	for _, part := range parts {
		diskInfo, _ := disk.Usage(part.Mountpoint)
		device := diskInfo.Path
		diskFree := toMbAndGb(diskInfo.Free)
		diskTotal := toMbAndGb(diskInfo.Total)

		diskPercent, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", diskInfo.UsedPercent), 64)
		s := fmt.Sprintf("\x1B[4;32;40m%s\t\t%s\t\t%s\t\t%v\x1B[0m\n", device, diskFree, diskTotal, diskPercent)
		io.WriteString(os.Stdout, s)

	}
}
