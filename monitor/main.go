package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-gomail/gomail"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/spf13/viper"
)

/*
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
*/
type EmailSubject struct {
	*SendMessage `yaml: "sendmessage"`
	*Message     `yaml:"message"`
	*Diskmessage `yaml:"diskmessage"`
}

type SendMessage struct {
	Fromemail  string   `yaml: "fromemail"`
	Smtpserver string   `yaml: "smtpserver"`
	Smtpport   int      `yaml: "smtpport"`
	Password   string   `yaml: "password"`
	Toemail    []string `yaml: "toemail"`
}

type Message struct {
	Subject   string `yaml: "subject"`
	File      string `yaml: "file"`
	Body_type string `yaml: "bodytype"`
	Body      string `yaml: "body"`
}

type Diskmessage struct {
	Partition string `yaml: "partition"`
}

type MessageTemp struct {
	DiskDevice  string
	DiskFree    float64
	DiskUsed    float64
	DiskPercent float64
}

var showDetail string

func initConfig() *EmailSubject {
	viper.SetConfigFile("/etc/config.yml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("no this file %w \n", err))
	}
	var senda *EmailSubject
	err = viper.Unmarshal(&senda)
	if err != nil {
		fmt.Println("over")
	}
	return senda
}

func sendEmail(femail, smtpserver, smtppassword, esubject string, smtpport int, attachfile, body string, toemails []string) {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", femail, "发件人")
	//m.SetHeader("To", m.FormatAddress(femail..., "收件人"))
	m.SetHeader("To", toemails...)
	m.SetHeader("Subject", esubject)
	m.SetBody("text/html", body)
	//m.SetBody("Content-Type: text/html; charset=utf-8", body)
	if len(attachfile) == 0 {
		fmt.Println("没有附件")
	} else {
		m.Attach(attachfile)
	}
	//m.Attach(attachfile)
	d := gomail.NewDialer(smtpserver, smtpport, femail, smtppassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	err := d.DialAndSend(m)
	if err != nil {
		fmt.Println("send error", err.Error())
		return
	}
	fmt.Println("邮件发送完成")
}

func toMbAndGb(size uint64) string {
	if size < 1024 {
		return strconv.FormatInt(int64(size), 10) + " Byte"
	} else if size >= 1024 && size < 1024000 {
		return strconv.FormatInt(int64(size/1024), 10) + " KB"
	} else if size >= 1024000 && size < 1048576 {
		return strconv.FormatInt(int64(size/1024/1024), 10) + " MB"
	} else {
		return strconv.FormatInt(int64(size/1024/1024/1024), 10) + " GB"
	}
}

func getDisk() {
	parts, err := disk.Partitions(true)
	if err != nil {
		fmt.Errorf("读取磁盘分区error:%s", err)
	}
	for _, part := range parts {
		z_DiskInfo, _ := disk.Usage(part.Mountpoint)
		if z_DiskInfo.Free <= 10737418240 { //磁盘小于10G进行邮件告警   5368709120
			showDetail = showDetail + `<h2>磁盘信息：</h2>
	<p>
		磁盘挂载点	|磁盘总量	|磁盘使用量	|磁盘剩余量	|磁盘使用百分比(%)<hr>`
			showDetail = showDetail + `
		` + z_DiskInfo.Path + `		` + toMbAndGb(z_DiskInfo.Total) + `		` + toMbAndGb(z_DiskInfo.Used) + `		` + toMbAndGb(z_DiskInfo.Free) + `		` + strconv.FormatUint(uint64(z_DiskInfo.UsedPercent), 10) + `
		<hr></p>`
		}
	}
}

func getDiskLinux() {
	shellScript := "df -h|awk -v threshold=80 -F\" \" '{if(NR>1) {if(substr($5,length($5-1)) > threshold) print $0}else{print $0}}'"
	cmd := exec.Command("bash", "-c", shellScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Errorf("执行命令错误：%s", err)
	} else {
		if strings.Contains(showDetail, "G") {
			showDetail = showDetail + `<h2>磁盘信息：</h2>
		<p>
			` + string(out) + `
		<hr></p>
		`
		}

	}
}
func getMem() {
	memInfo, _ := mem.VirtualMemory()
	if memInfo.UsedPercent > 90 {
		showDetail = showDetail + `<h2>内存信息</h2>
	<p>
		内存总量	|可用内存	|使用内存	|内存使用百分比（%）<hr>
		` + toMbAndGb(memInfo.Total) + `	` + toMbAndGb(memInfo.Available) + `	` + toMbAndGb(memInfo.Used) + `	` + strconv.FormatUint(uint64(memInfo.UsedPercent), 10) + `
	<hr></p>
	`
	}
}

func getCpuLoad() string {
	num, _ := cpu.Counts(false)
	info, _ := load.Avg()
	if info.Load1 >= float64(num) {
		showDetail = showDetail + `<h2>CPU负载信息</h2>
	<p>
		1min	|5min	|15min<hr>
		` + strconv.FormatFloat(info.Load1, 'f', -1, 64) + `		` + strconv.FormatFloat(info.Load5, 'f', -1, 64) + `		` + strconv.FormatFloat(info.Load15, 'f', -1, 64) + `
		<hr></p>
	`
	}
	return showDetail
}

func GetLocalIP() (ip string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, addr := range addrs {
		ipAddr, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if ipAddr.IP.IsLoopback() {
			continue
		}
		if !ipAddr.IP.IsGlobalUnicast() {
			continue
		}
		return ipAddr.IP.String(), nil
	}
	return
}

func main() {
	var message_model string
	var help string
	var subject string
	var version bool
	var attach string
	var example string
	myip, err := GetLocalIP()
	if err != nil {
		fmt.Errorf("获取ip地址失败,error:%s", err)
	}
	//获取磁盘信息
	if runtime.GOOS == "linux" {
		getDiskLinux()
	} else {
		getDisk()
	}
	getMem()
	resultMessage := getCpuLoad()
	//获取配置文件信息
	get_message := initConfig()
	//disk_message := getDisk(get_message.Partition)
	//fmt.Println(disk_message.Total)
	flag.StringVar(&help, "h", "", "获取帮助信息")
	flag.StringVar(&message_model, "m", resultMessage, "邮件发送内容，与-s一起使用")
	flag.StringVar(&subject, "s", "", "邮件主题")
	flag.BoolVar(&version, "v", false, "查询版本信息")
	flag.StringVar(&attach, "f", "", "添加邮件附件")
	flag.StringVar(&example, "e", "", "程序示例")
	flag.Parse()
	//fmt.Println(get_message.Toemail)
	if version {
		fmt.Printf(`Version:1.0`)
		return
	} else if example == "example" {
		if runtime.GOOS == "windows" {
			fmt.Print(`.\monitor.exe -m "<h1>磁盘告警邮件:xxxxx磁盘超过90%，请及时清理</h1>" -s "邮件告警" -f "C:\Users\EDZ\Desktop\xxxxxx.pdf"`)
		} else {
			fmt.Print(`./monitor -m "<h1>磁盘告警邮件:xxxxx磁盘超过90%，请及时清理</h1>" -s "邮件告警" -f "/home/xxxxxx.pdf"`)
		}

	} else if message_model != "" && subject != "" {
		//调用发送邮件
		sendEmail(get_message.Fromemail, get_message.Smtpserver, get_message.Password, myip+subject, get_message.Smtpport, attach, message_model, get_message.Toemail)
	} else {
		fmt.Printf("%s  [info] 磁盘、负载、内存指标正常\n", time.Now().Format("2006-01-02 15:04:05"))
	}

}
