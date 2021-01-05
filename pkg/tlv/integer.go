package tlv

import "strconv"

//
type IntegerVal int64

// GeneratePayload takes in the identifier and the value of a resource, and converts them into an LWM2M paylaod.
func (i IntegerVal) GeneratePayload(identifier IntegerVal) []byte {
	header := generateHeader(identifier, i)
	value := bitStreamToByteSlice(i.toBitString())
	return append(header, value...)
}

func (i IntegerVal) getByteLength() int64 {
	s1 := i.toBitString()
	s2 := bitStreamToByteSlice(s1)
	return int64(len(s2))
}

func (i IntegerVal) toBitString() string {
	s := strconv.FormatInt(int64(i), 2)
	if t := len(s) % 8; t != 0 {
		for i := 0; i < 8-t; i++ {
			s = "0" + s
		}
	}
	return s
}

func (i IntegerVal) toBitStringWithZeros(zeros int) string {
	s := strconv.FormatInt(int64(i), 2)
	if t := len(s) % zeros; t != 0 {
		for i := 0; i < zeros-t; i++ {
			s = "0" + s
		}
	}
	return s
}

func (i IntegerVal) toBitStringWithoutZeros() string {
	s := strconv.FormatInt(int64(i), 2)
	return s
}
