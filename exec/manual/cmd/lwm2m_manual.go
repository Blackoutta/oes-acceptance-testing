package cmd

import (
	"fmt"
	"oes-acceptance-testing/v1/pkg/nbclient"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var lwm2mDesc = []string{
	"lwm2m设备模拟器，配置文件格式见./configs/manual/lwm2m_manual.yaml",
	"不支持自定义上行payload, 上行payload目前为固定, 无需定义物模型",
}

var lwm2mConfS = &lwm2mconf{}
var lwm2mCmd = &cobra.Command{
	Use:   "lwm2m",
	Short: "模拟lwm2m设备",
	Long:  strings.Join(lwm2mDesc, "\n"),
	Run:   RunLwm2mSim,
}

func init() {
	rootCmd.AddCommand(lwm2mCmd)
	v := newSetting("lwm2m_manual")
	v.Unmarshal(&lwm2mConfS)
}

func RunLwm2mSim(cmd *cobra.Command, args []string) {
	fmt.Println(lwm2mConfS.Host)
	//启动客户端
	nc := nbclient.NbiotClient{
		Addr:     lwm2mConfS.Host,
		IMEI:     lwm2mConfS.Imei,
		Interval: time.Duration(lwm2mConfS.Interval),
	}

	go func() {
		nc.Boot()
	}()

	wg.Add(1)
	wg.Wait()
}
