package tlv

import (
	"strconv"
)

type tlvHeader struct {
	TypeOfIdentifier  IntegerVal
	LengthOfIdentifer IntegerVal
	LengthOfLength    IntegerVal // value长度小于8个字节时，置为0；否则置为1，且将原length占用的3个位置为0，然后在identifer后面放上新的length
	Length            IntegerVal //length of value
}

type TLVValue interface {
	getByteLength() int64
}

func generateHeader(identifier IntegerVal, value TLVValue) []byte {
	var th tlvHeader
	th.TypeOfIdentifier = 3
	valLen := value.getByteLength()

	var finalHeader []byte
	//如果value长度小于8个字节
	if valLen <= 7 {
		if identifier.getByteLength() == 1 {
			th.LengthOfIdentifer = 0
		}
		if identifier.getByteLength() == 2 {
			th.LengthOfIdentifer = 1
		}

		th.LengthOfLength = 0
		th.Length = IntegerVal(valLen)

		// 拼装LWM2M报文 - 首先转为bitstream
		var msg string
		msg += th.TypeOfIdentifier.toBitStringWithoutZeros()
		msg += th.LengthOfIdentifer.toBitStringWithoutZeros()
		msg += th.LengthOfLength.toBitStringWithZeros(2)
		msg += th.Length.toBitStringWithZeros(3)
		msg += identifier.toBitString()

		// // 接着从bitsteam转byte slice
		finalHeader = bitStreamToByteSlice(msg)
	}

	//如果value长度大于8个字节
	if valLen > 7 {
		if identifier.getByteLength() == 1 {
			th.LengthOfIdentifer = 0
		}
		if identifier.getByteLength() == 2 {
			th.LengthOfIdentifer = 1
		}

		th.LengthOfLength = 1
		th.Length = 0

		// 拼装LWM2M报文 - 首先转为bitstream
		var msg string
		msg += th.TypeOfIdentifier.toBitStringWithoutZeros()
		msg += th.LengthOfIdentifer.toBitStringWithoutZeros()
		msg += th.LengthOfLength.toBitStringWithZeros(2)
		msg += th.Length.toBitStringWithZeros(3)
		msg += identifier.toBitString()
		msg += IntegerVal(valLen).toBitString()

		// // 接着从bitsteam转byte slice
		finalHeader = bitStreamToByteSlice(msg)

	}
	return finalHeader
}

func bitStreamToByteSlice(b string) []byte {
	var out []byte
	var str string

	for i := len(b); i > 0; i -= 8 {
		if i-8 < 0 {
			str = string(b[0:i])
		} else {
			str = string(b[i-8 : i])
		}
		v, err := strconv.ParseUint(str, 2, 8)
		if err != nil {
			panic(err)
		}
		out = append([]byte{byte(v)}, out...)
	}
	return out
}
