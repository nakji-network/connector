package common

import (
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
)

// milliseconds
func UnixToTimestampPb(unixtime int64) *timestamp.Timestamp {
	return &timestamp.Timestamp{
		Seconds: unixtime / 1000,
		Nanos:   int32(unixtime % 1000 * 1e6),
	}
}

func DecodeBigIntArray(arrayField []*big.Int) [][]byte {
	newSlice := make([][]byte, len(arrayField))

	for i, value := range arrayField {
		newSlice[i] = value.Bytes()
	}

	return newSlice
}

func DecodeAddressArray(arrayField []ethcommon.Address) [][]byte {
	newSlice := make([][]byte, len(arrayField))

	for i, value := range arrayField {
		newSlice[i] = value.Bytes()
	}

	return newSlice
}

func DecodeUint8Array(arrayField []uint8) []uint32 {
	newSlice := make([]uint32, len(arrayField))

	for i, value := range arrayField {
		newSlice[i] = uint32(value)
	}

	return newSlice
}

func DecodeUint16Array(arrayField []uint16) []uint32 {
	newSlice := make([]uint32, len(arrayField))

	for i, value := range arrayField {
		newSlice[i] = uint32(value)
	}

	return newSlice
}

func DecodeInt8Array(arrayField []int8) []int32 {
	newSlice := make([]int32, len(arrayField))

	for i, value := range arrayField {
		newSlice[i] = int32(value)
	}

	return newSlice
}

func DecodeInt16Array(arrayField []int16) []int32 {
	newSlice := make([]int32, len(arrayField))

	for i, value := range arrayField {
		newSlice[i] = int32(value)
	}

	return newSlice
}
