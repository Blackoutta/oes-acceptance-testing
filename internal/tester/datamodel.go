package tester

import (
	"fmt"
	"math"
	"time"
)

type Datapoint struct {
	Value interface{} `json:"value"`
	// Time  int64       `json:"time,omitemtpy"`
}

func NewDataModel() map[string]Datapoint {
	data := make(map[string]Datapoint)
	time := time.Now().Unix() * 1000

	data["someDouble"] = Datapoint{
		Value: 999999.999999999,
	}
	data["someFloat"] = Datapoint{
		Value: 999.999,
	}
	data["someInteger"] = Datapoint{
		Value: math.MaxInt32,
	}
	data["someLong"] = Datapoint{
		Value: 999999999999999,
	}
	data["someBoolean"] = Datapoint{
		Value: false,
	}
	data["someDate"] = Datapoint{
		Value: time,
	}
	data["someString"] = Datapoint{
		Value: "test string",
	}
	data["someEnum"] = Datapoint{
		Value: 20,
	}

	return data
}

func (t *Tester) CreateProperties() {
	var err error
	// 添加功能double
	cpr := CreateDmpPropertyPl{
		Name:       "someDouble",
		Identifier: "someDouble",
		AccessMode: 2,
		Type:       DATATYPE_FLOAT64,
		Unit:       "some_unit",
		Minimum:    float64(-999999.999999999),
		Maximum:    float64(999999.999999999),
		Special: Special{
			Step: 0.000000001,
		},
	}
	t.Record, err = t.CreateDmpProperty(t.Params.ProductID, cpr)
	t.Must(err)

	t.AssertTrue("添加功能double", t.Record.IsSuccess())

	// 添加功能string
	strp := CreateDmpPropertyPl{
		Name:       "someString",
		Identifier: "someString",
		AccessMode: 2,
		Type:       DATATYPE_STRING,
		Unit:       "some_unit",
		Special: Special{
			Length: 2048,
		},
	}
	t.Record, err = t.CreateDmpProperty(t.Params.ProductID, strp)
	t.Must(err)

	t.AssertTrue("添加功能string", t.Record.IsSuccess())

	// 添加功能integer
	int32p := CreateDmpPropertyPl{
		Name:       "someInteger",
		Identifier: "someInteger",
		AccessMode: 2,
		Type:       DATATYPE_INT32,
		Unit:       "some_unit",
		Minimum:    math.MinInt32,
		Maximum:    math.MaxInt32,
		Special: Special{
			Step: 1,
		},
	}
	t.Record, err = t.CreateDmpProperty(t.Params.ProductID, int32p)
	t.Must(err)

	t.AssertTrue("添加功能int32", t.Record.IsSuccess())

	// 添加Long
	int64p := CreateDmpPropertyPl{
		Name:       "someLong",
		Identifier: "someLong",
		AccessMode: 2,
		Type:       DATATYPE_INT64,
		Unit:       "some_unit",
		Minimum:    -999999999999999,
		Maximum:    999999999999999,
		Special: Special{
			Step: 1,
		},
	}
	t.Record, err = t.CreateDmpProperty(t.Params.ProductID, int64p)
	t.Must(err)

	t.AssertTrue("添加功能int64", t.Record.IsSuccess())

	// 添加Float32
	float32p := CreateDmpPropertyPl{
		Name:       "someFloat",
		Identifier: "someFloat",
		AccessMode: 2,
		Type:       DATATYPE_FLOAT32,
		Unit:       "some_unit",
		Minimum:    -999.999,
		Maximum:    999.999,
		Special: Special{
			Step: 0.001,
		},
	}
	t.Record, err = t.CreateDmpProperty(t.Params.ProductID, float32p)
	t.Must(err)

	t.AssertTrue("添加功能float32", t.Record.IsSuccess())

	// 添加boolean
	boolp := CreateDmpPropertyPl{
		Name:       "someBoolean",
		Identifier: "someBoolean",
		AccessMode: 2,
		Type:       DATATYPE_BOOLEAN,
		Unit:       "some_unit",
		Special: Special{
			EnumArrays: []EnumArray{
				{
					Key:      0,
					Describe: "false",
				},
				{
					Key:      1,
					Describe: "true",
				},
			},
		},
	}

	t.Record, err = t.CreateDmpProperty(t.Params.ProductID, boolp)
	t.Must(err)

	t.AssertTrue("添加功能boolean", t.Record.IsSuccess())

	// 添加Enum
	enumArr := make([]EnumArray, 0, 20)
	for i := 1; i <= 20; i++ {
		enumArr = append(enumArr, EnumArray{
			Key:      i,
			Describe: fmt.Sprintf("val_%v", i),
		})
	}

	enump := CreateDmpPropertyPl{
		Name:       "someEnum",
		Identifier: "someEnum",
		AccessMode: 2,
		Type:       DATATYPE_ENUM,
		Unit:       "some_unit",
		Special: Special{
			EnumArrays: enumArr,
		},
	}

	t.Record, err = t.CreateDmpProperty(t.Params.ProductID, enump)
	t.Must(err)

	t.AssertTrue("添加功能Enum", t.Record.IsSuccess())

	// 添加Date
	datep := CreateDmpPropertyPl{
		Name:       "someDate",
		Identifier: "someDate",
		AccessMode: 2,
		Type:       DATATYPE_DATE,
		Unit:       "some_unit",
	}
	t.Record, err = t.CreateDmpProperty(t.Params.ProductID, datep)
	t.Must(err)

	t.AssertTrue("添加功能date", t.Record.IsSuccess())
}
