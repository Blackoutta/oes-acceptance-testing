package tester

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Tester struct {
	SuiteName   string
	SuiteFailed bool
	Teardown    bool
	Logger      *log.Logger
	Params      Params
	Record      Record
	Request     *Request
	Client      *http.Client
}

func (t *Tester) Must(err error) {
	if err != nil {
		t.Logger.Println(err)
		t.SuiteFailed = true
	}
}

func (t *Tester) DisplayResult(resultChan chan TestResult) {
	if t.SuiteFailed == true {
		t.Logger.Printf("======= %v 测试失败！/(ㄒoㄒ)/，请检查日志定位问题 =======", t.SuiteName)
	} else {
		t.Logger.Printf("======= %v 测试通过！O(∩_∩)O =======", t.SuiteName)
	}

	r := TestResult{
		SuiteName: t.SuiteName,
		Success:   !t.SuiteFailed,
	}
	resultChan <- r
}

func (t *Tester) FailTest(msg string) {
	t.SuiteFailed = true
	t.Logger.Println(msg)
}

func (t *Tester) Recover() {
	if a := recover(); a != nil {
		t.SuiteFailed = true
		t.Logger.Printf("发生panic: %v, 测试失败，程序将开始TEARDOWN!", a)
	}
}

func HandleResult(resultChan chan TestResult, scenNum int) {
	var result string
	var exitCode int

	result += fmt.Sprintf("%-10v%-30v%-10v\n", "ID", "Suite", "Success?")

	for i := 0; i < scenNum; i++ {
		r := <-resultChan
		if r.Success == false {
			exitCode = 1
		}

		suiteName := r.SuiteName
		var chCount int
		for _, c := range suiteName {
			if c > 11000 {
				chCount++
			}
		}
		pattern := strings.Replace("%-10v%-30v%-10v\n", "30", strconv.Itoa(30-chCount), -1)

		result += fmt.Sprintf(pattern, i+1, r.SuiteName, r.Success)
	}

	fmt.Print(result)
	os.Exit(exitCode)
}

type TestResult struct {
	SuiteName string
	Success   bool
}

type ScenarioFn func(chan TestResult, bool)

func SetOnhold(onHold *bool) {
	flag.BoolVar(onHold, "h", false, "用来决定测试程序执行完毕后，是否保留所有资源并让设备持续在线")
	flag.Parse()
}
