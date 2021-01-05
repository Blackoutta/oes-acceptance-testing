package cmd

import (
	"oes-acceptance-testing/v1/pkg/modbusSlave"
	"strings"

	"github.com/spf13/cobra"
)

var modbusDesc = []string{
	"modbus设备模拟器，配置文件格式见./configs/manual/modbus_manual.yaml",
	"不支持自定义上行payload, 上行需要固定的物模型文件，见./data/model/modbus_model.json",
}

var modbusConfS = &modbusconf{}
var modbusCmd = &cobra.Command{
	Use:   "modbus",
	Short: "模拟modbus设备",
	Long:  strings.Join(modbusDesc, "\n"),
	Run:   RunModbusSim,
}

func init() {
	rootCmd.AddCommand(modbusCmd)
	v := newSetting("modbus_manual")
	v.Unmarshal(&modbusConfS)
}

func RunModbusSim(cmd *cobra.Command, args []string) {
	s := modbusSlave.ModbusSlave{
		ProductID:        modbusConfS.Pid,
		DeviceID:         modbusConfS.Did,
		DeviceSecret:     modbusConfS.DevSecret,
		SlaveID:          modbusConfS.SlaveID,
		ModbusMasterAddr: modbusConfS.Host,
	}

	go func() {
		s.Boot()
	}()
	wg.Add(1)
	wg.Wait()
}
