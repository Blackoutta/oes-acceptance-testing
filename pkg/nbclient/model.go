package nbclient

import "oes-acceptance-testing/v1/pkg/tlv"

const (
	RESOURCE_TEMPERATURE = 3303
	RESOURCE_SETPOINT    = 3308
)

const (
	TEMPERATURE_SENSOR_VALUE tlv.IntegerVal = 5700
	SETPOINT_SETPOINT_VALUE  tlv.IntegerVal = 5900
)

type resource interface {
	getFloatValue() tlv.FloatVal
}

type temperature struct {
	sensorValue tlv.FloatVal
}

func (t *temperature) getFloatValue() tlv.FloatVal {
	return t.sensorValue
}

type setPoint struct {
	setPointValue tlv.FloatVal
}

func (t *setPoint) getFloatValue() tlv.FloatVal {
	return t.setPointValue
}
