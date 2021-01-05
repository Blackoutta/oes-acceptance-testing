package cmd

import (
	"fmt"
	"oes-acceptance-testing/v1/global"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var wg sync.WaitGroup
var rootCmd = &cobra.Command{
	Use:   "oes-sim",
	Short: "OES设备模拟器",
	Long:  "OneNET智能边缘计算套件设备模拟器，支持协议：mqtt, mqtt透传，tcp透传，modbus rtu over tcp, lwm2m, http",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func newSetting(name string) *viper.Viper {
	v := viper.New()
	v.SetConfigName(name)
	v.SetConfigType("yaml")
	v.AddConfigPath(global.WorkDir() + "/configs/manual/")
	v.AddConfigPath("./configs/manual/")
	v.AddConfigPath("./")
	v.AddConfigPath(global.WorkDir())
	v.AddConfigPath(global.WorkDir() + "/configs/")
	v.AddConfigPath("./configs/")
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("加载配置文件时发生错误: %v", err))
	}
	return v
}
