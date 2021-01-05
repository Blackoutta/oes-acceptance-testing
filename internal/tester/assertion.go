package tester

import (
	"fmt"
	"strings"
)

const (
	passSubFix = "\t......Pass!\n"
	failSubfix = "\t......Fail!\n"
)

func (t *Tester) AssertTrue(title string, boolean bool) {
	if boolean != true {
		t.SuiteFailed = true
		astResult := fmt.Sprintf("Expect value to equal true, got %v", boolean)
		t.Logger.Printf(title + failSubfix)
		t.printErrorInfo(t.Record, astResult)
		return
	}
	t.Logger.Printf(title + passSubFix)
}

func (t *Tester) AssertFloat64(title string, expect float64, actual float64) {
	if expect != actual {
		t.SuiteFailed = true
		astResult := fmt.Sprintf("Expect value to equal %v, got %v", expect, actual)
		t.Logger.Printf(title + failSubfix)
		t.printErrorInfo(t.Record, astResult)
		return
	}
	t.Logger.Printf(title + passSubFix)
}

func (t *Tester) AssertInteger(title string, expect int, actual int) {
	if expect != actual {
		t.SuiteFailed = true
		astResult := fmt.Sprintf("Expect value to equal %v, got %v", expect, actual)
		t.Logger.Printf(title + failSubfix)
		t.printErrorInfo(t.Record, astResult)
		return
	}
	t.Logger.Printf(title + passSubFix)
}

func (t *Tester) AssertIntegerGreaterThan(title string, expect int, actual int) {
	if expect >= actual {
		t.SuiteFailed = true
		astResult := fmt.Sprintf("Expect value to be greater than %v, got %v", expect, actual)
		t.Logger.Printf(title + failSubfix)
		t.printErrorInfo(t.Record, astResult)
		return
	}
	t.Logger.Printf(title + passSubFix)
}

func (t *Tester) AssertStringEqual(title string, expect string, actual string) {
	if expect != actual {
		t.SuiteFailed = true
		astResult := fmt.Sprintf("Expect value to equal %v, got %v", expect, actual)
		t.Logger.Printf(title + failSubfix)
		t.printErrorInfo(t.Record, astResult)
		return
	}
	t.Logger.Printf(title + passSubFix)
}

func (t *Tester) AssertStringContains(title string, origin string, contain string) {
	if !strings.Contains(origin, contain) {
		t.SuiteFailed = true
		astResult := fmt.Sprintf("Expect %v to contain %v\n", origin, contain)
		t.Logger.Printf(title + failSubfix)
		t.printErrorInfo(t.Record, astResult)
		return
	}
	t.Logger.Printf(title + passSubFix)
}

func (t *Tester) printErrorInfo(rec Record, astResult string) {
	t.Logger.Println("============================== ERROR INFO START ==============================")
	t.Logger.Printf("\nSUCCESS: \t%v\nURL: \t\t%v\nMETHOD: \t%v\nBODY: \n%v\nCODE: \t\t%v\nMSG: \t\t%v\nAssertion Result: \t\t%v\n",
		rec.Response.IsSuccess(),
		rec.ReqURL,
		rec.ReqMethod,
		rec.ReqBody,
		rec.Response.GetCode(),
		rec.Response.GetMsg(),
		astResult)
	t.Logger.Println("============================== ERROR INFO END ==============================")
}
