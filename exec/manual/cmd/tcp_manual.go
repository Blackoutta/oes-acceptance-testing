package cmd

import (
	"log"
	"oes-acceptance-testing/v1/pkg/tcpClient"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var tcpDesc = []string{
	"tcp透传设备模拟器，配置文件格式见./configs/manual/tcp_manual.yaml",
	"支持自定义上行payload，上行payload文件见./data/payload/tcp.json",
	"上行数据前需通过页面配置解析脚本, 脚本可在页面上的'产品-数据解析-模板'中找到，或在本仓的./data/tcpscript文件中找到",
}

var tcpConfS = &conf{}
var tcpCmd = &cobra.Command{
	Use:   "tcp",
	Short: "模拟tcp设备",
	Long:  strings.Join(tcpDesc, "\n"),
	Run:   RunTcpSim,
}

func init() {
	rootCmd.AddCommand(tcpCmd)
	v := newSetting("tcp_manual")
	v.Unmarshal(&tcpConfS)
}

func RunTcpSim(cmd *cobra.Command, args []string) {
	c := tcpClient.TCPClient{
		ProductID: tcpConfS.Pid,
		DeviceID:  tcpConfS.Did,
		DevSecret: tcpConfS.DevSecret,
	}

	err := c.Connect(tcpConfS.Host)
	if err != nil {
		log.Fatalf("连接tcp接入机时发生错误: %v", err)
	}

	time.Sleep(time.Second)

	go c.HandleCommand()
	go c.UpwardData(tcpConfS.PayloadFile, time.Duration(tcpConfS.Interval))
	go c.HeartBeat()
	wg.Add(1)
	wg.Wait()
}
