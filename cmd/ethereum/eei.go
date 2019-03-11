package main

import (
	"encoding/binary"
	"fmt"
	"os"

	"reflect"
)

func Caca() {
	fmt.Println("caca")
}

func importGetCallDataSize() uint64
func importUseGas() uint64
func importCallDataCopy() uint64
func importFinish() uint64
func importGrowMemoryHandler() uint64

var callData = []byte("coucou1")
var result []byte

func UseGas(a int64) {
	//precompile.contract.UseGas(uint64(a))
	//panic(fmt.Sprintf("use gas %d", a))
}

func CallDataCopyHelper(r, d, l uintptr) uintptr {
	return memAddr(callData) + d

	// panic(fmt.Sprintf("%d %d %d", r, d, l))
	// if l > 0 {
	// Unlike regular EEI functions, the gas is not charged at this
	// time but I'm leaving that code here for future reference.
	// in.gasAccounting(GasCostVeryLow + GasCostCopy*(uint64(l+31)>>5))
	//		p.WriteAt(precompile.input[d:d+l], int64(r))
	// }
}

func GetCallDataSize() int32 {
	return int32(len(callData))
}

func FinishHelper(l int32) uintptr {
	result = make([]byte, l)
	for i := range result {
		result[i] = 0x55
	}
	return memAddr(result)
}

func revert(d, l int32) {
	panic("revert")
	os.Exit(1)
	//	fmt.Println("terminate");
	//	precompile.retData = make([]byte, int64(l))
	//	p.ReadAt(precompile.retData, int64(d))
	//	p.Terminate()
}

func init() {
	importVector = make([]byte, (5+4)*8)
	binary.LittleEndian.PutUint64(importVector[0:], importUseGas())
	binary.LittleEndian.PutUint64(importVector[8:], importCallDataCopy())
	binary.LittleEndian.PutUint64(importVector[16:], importGetCallDataSize())
	binary.LittleEndian.PutUint64(importVector[24:], importFinish())
	binary.LittleEndian.PutUint64(importVector[32:], uint64(reflect.ValueOf(revert).Pointer()))
	importFuncs["useGas"] = importFunc{-9, 1}
	importFuncs["callDataCopy"] = importFunc{-8, 3}
	importFuncs["getCallDataSize"] = importFunc{-7, 0}
	importFuncs["finish"] = importFunc{-6, 2}
	importFuncs["revert"] = importFunc{-5, 2}

	binary.LittleEndian.PutUint64(importVector[56:], importGrowMemoryHandler())
}

func setImportVectorCurrentMemory(size int) {
	binary.LittleEndian.PutUint64(importVector[40:], uint64(size))
}
