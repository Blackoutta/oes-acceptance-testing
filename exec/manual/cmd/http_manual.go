package cmd

import (
	"io/ioutil"
	"log"
	"oes-acceptance-testing/v1/pkg/httpClient"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var httpDesc = []string{
	"http设备模拟器，配置文件格式见./configs/manual/http_manual.yaml",
	"支持自定义上行payload，上行payload文件见./data/payload/http.json",
}

var httpConfS = &conf{}
var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "模拟http设备",
	Long:  strings.Join(httpDesc, "\n"),
	Run:   RunHttpSim,
}

func init() {
	rootCmd.AddCommand(httpCmd)
	v := newSetting("http_manual")
	v.Unmarshal(&httpConfS)
}

func RunHttpSim(cmd *cobra.Command, args []string) {
	oc := httpClient.NewOesHttpClient(httpConfS.Pid, httpConfS.Did, httpConfS.DevSecret, httpConfS.Host)

	// 上行数据
	go func() {
		for {
			pl, err := ioutil.ReadFile(httpConfS.PayloadFile)
			r1 := oc.NewUpwardRequest(httpConfS.Host, string(pl))

			resp, err := oc.Client.Do(r1)
			if err != nil {
				log.Fatalf("error after sending upward request: %v", err)
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("已上行数据: %v", string(pl))
			if resp.StatusCode != 200 {
				log.Println(string(body))
				log.Fatalf("上行请求发送后响应码不为200，测试失败。")
			}
			time.Sleep(time.Duration(httpConfS.Interval) * time.Second)
		}
	}()
	wg.Add(1)
	wg.Wait()
}
