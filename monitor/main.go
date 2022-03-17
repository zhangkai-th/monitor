package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/go-gomail/gomail"
	"github.com/shirou/gopsutil/disk"
	"github.com/spf13/viper"
)

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

func initConfig() *EmailSubject {
	viper.SetConfigFile("../config.yml")
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
	//m.Attach(attachfile)
	d := gomail.NewDialer(smtpserver, smtpport, femail, smtppassword)
	err := d.DialAndSend(m)
	if err != nil {
		fmt.Println("send error", err.Error())
		return
	}
	fmt.Println("邮件发送完成")
}

func getDisk(diskDriver string) *disk.UsageStat {
	d, _ := disk.Usage(diskDriver)
	return d
}

func main() {
	var message_model string
	var help string
	var subject string
	var version bool
	var attach string
	var example string

	//获取配置文件信息
	get_message := initConfig()
	disk_message := getDisk(get_message.Partition)
	fmt.Println(disk_message.Total)
	flag.StringVar(&help, "h", "", "获取帮助信息")
	flag.StringVar(&message_model, "m", "", "邮件发送内容，与-s一起使用")
	flag.StringVar(&subject, "s", "", "邮件主题")
	flag.BoolVar(&version, "v", false, "查询版本信息")
	flag.StringVar(&attach, "f", "", "添加邮件附件")
	flag.StringVar(&example, "e", "", "程序示例")
	flag.Parse()
	fmt.Println(get_message.Toemail)
	if version {
		fmt.Printf(`Version:1.0`)
		return
	} else if example == "example" {
		if runtime.GOOS == "windows" {
			fmt.Print(`.\monitor.exe -m "<h1>磁盘告警邮件:xxxxx磁盘超过90%，请及时清理</h1>" -s "邮件告警" -f "C:\Users\EDZ\Desktop\xxxxxx.pdf"`)
		} else {
			fmt.Print(`./monitor.exe -m "<h1>磁盘告警邮件:xxxxx磁盘超过90%，请及时清理</h1>" -s "邮件告警" -f "/home/xxxxxx.pdf"`)
		}

	} else if message_model != "" && subject != "" {
		//调用发送邮件
		sendEmail(get_message.Fromemail, get_message.Smtpserver, get_message.Password, subject, get_message.Smtpport, attach, message_model, get_message.Toemail)
	}

}
