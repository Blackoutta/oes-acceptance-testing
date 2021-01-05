package tlv

// StringVal represents an lwm2m string value
type StringVal string

func (i StringVal) getByteLength() int64 {
	bs := []byte(i)
	return int64(len(bs))
}

// GeneratePayload generates an lwm2m tlv payload from a string value
func (i StringVal) GeneratePayload(identifier IntegerVal) []byte {
	header := generateHeader(identifier, i)
	return append(header, []byte(i)...)
}
