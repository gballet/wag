package main

import (
	"encoding/binary"
	"fmt"
)

func Caca() {
	fmt.Println("caca")
}

func importGetCallDataSize() uint64
func importUseGas() uint64
func importCallDataCopy() uint64
func importFinish() uint64
func importRevert() uint64
func importGrowMemoryHandler() uint64

var callData = []byte("coucou1")
var result []byte

func init() {
	importVector = make([]byte, (5+4)*8)
	binary.LittleEndian.PutUint64(importVector[0:], importUseGas())
	binary.LittleEndian.PutUint64(importVector[8:], importCallDataCopy())
	binary.LittleEndian.PutUint64(importVector[16:], importGetCallDataSize())
	binary.LittleEndian.PutUint64(importVector[24:], importFinish())
	binary.LittleEndian.PutUint64(importVector[32:], importRevert())
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
