package tlv

import (
	"encoding/binary"
	"math"
)

type FloatVal float64

// GeneratePayload takes in the identifier and the value of a resource, and converts them into an LWM2M paylaod.
func (i FloatVal) GeneratePayload(identifier IntegerVal) []byte {
	header := generateHeader(identifier, i)
	value := i.GetBytes()
	payload := append(header, value...)
	return payload
}

func (i FloatVal) getByteLength() int64 {
	return int64(len(i.GetBytes()))
}

func (i FloatVal) GetBytes() []byte {
	bits := math.Float64bits(float64(i))
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, bits)
	return bytes
}
