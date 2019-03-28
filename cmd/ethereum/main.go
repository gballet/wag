// Copyright (c) 2018 Timo Savola. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Program wasys implements a standalone toy compiler and runtime.
package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"syscall"
	"unsafe"

	"github.com/tsavola/wag/internal/gen"

	"github.com/tsavola/wag"
	"github.com/tsavola/wag/buffer"
	"github.com/tsavola/wag/compile"
	"github.com/tsavola/wag/object/debug/dump"
	"github.com/tsavola/wag/wa"
)

const linearMemoryAddressSpace = 8 * 1024 * 1024 * 1024
const signalStackReserve = 8192

var (
	verbose = false
)

type importFunc struct {
	index  int
	params int
}

var (
	importFuncs  = make(map[string]importFunc)
	importVector []byte
)

type resolver struct{}

func (resolver) ResolveFunc(module, field string, sig wa.FuncType) (index int, err error) {
	if verbose {
		log.Printf("import %s%s", field, sig)
	}

	if module != "ethereum" {
		err = fmt.Errorf("import function's module is unknown: %s %s", module, field)
		return
	}

	i := importFuncs[field]
	if i.index == 0 {
		err = fmt.Errorf("import function not supported: %s", field)
		return
	}
	if len(sig.Params) != i.params {
		err = fmt.Errorf("%s: import function has wrong number of parameters: import signature has %d, syscall wrapper has %d", field, len(sig.Params), i.params)
		return
	}

	index = i.index
	return
}

func (resolver) ResolveGlobal(module, field string, t wa.Type) (init uint64, err error) {
	err = fmt.Errorf("imported global not supported: %s %s", module, field)
	return
}

func makeMem(size int, prot, extraFlags int) (mem []byte, err error) {
	if size > 0 {
		mem, err = syscall.Mmap(-1, 0, size, prot, syscall.MAP_PRIVATE|syscall.MAP_ANONYMOUS|extraFlags)
	}
	return
}

func memAddr(mem []byte) uintptr {
	return (*reflect.SliceHeader)(unsafe.Pointer(&mem)).Data
}

func alignSize(size, alignment int) int {
	return (size + (alignment - 1)) &^ (alignment - 1)
}

func GrowMemory(size int32) {

}

func main() {
	log.SetFlags(0)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] wasmfile\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	var (
		textSize  = compile.DefaultMaxTextSize
		stackSize = wa.PageSize
		entry     = "main"
		dumpText  = false
		inputData = "Hello, world!"
		inputHex  = ""
		startGas  = uint64(1000)
	)

	flag.BoolVar(&verbose, "v", verbose, "verbose logging")
	flag.IntVar(&textSize, "textsize", textSize, "maximum program text size")
	flag.IntVar(&stackSize, "stacksize", stackSize, "call stack size")
	flag.StringVar(&entry, "entry", entry, "function to run")
	flag.BoolVar(&dumpText, "dumptext", dumpText, "disassemble the generated code to stdout")
	flag.StringVar(&inputData, "input", inputData, "An (escaped) string representing the bytes of the input")
	flag.StringVar(&inputHex, "input-hex", inputHex, "A string containing the hex representation of the bytes of the input. Overrides -input")
	flag.Uint64Var(&startGas, "startgas", startGas, "Initial amount of gas to start the program with")
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	filename := flag.Arg(0)

	prog, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	progReader := bytes.NewReader(prog)

	vecSize := alignSize(len(importVector), os.Getpagesize())

	vecTextMem, err := makeMem(vecSize+textSize, syscall.PROT_READ|syscall.PROT_WRITE, 0)
	if err != nil {
		log.Fatal(err)
	}

	vecMem := vecTextMem[:vecSize]
	copy(vecMem[vecSize-len(importVector):], importVector)

	var input []byte
	if len(inputHex) != 0 {
		input, err = hex.DecodeString(inputHex)
		if err != nil {
			fmt.Println("Cannot decode hex string", inputHex)
			os.Exit(1)
		}
	} else {
		input = []byte(inputData)
	}

	contractData := make([]byte, 8+8+8+len(input) /* original rsp + gas + size + data */)
	binary.LittleEndian.PutUint64(contractData[8:], startGas)
	binary.LittleEndian.PutUint64(contractData[16:], uint64(len(input)))
	copy(contractData[24:], input)
	cdAddr := uint64(memAddr(contractData))
	binary.LittleEndian.PutUint64(vecTextMem[vecSize+gen.VectorOffsetGoStack:], cdAddr)

	textMem := vecTextMem[vecSize:]
	textAddr := memAddr(textMem)
	textBuf := buffer.NewStatic(textMem[:0], len(textMem))

	config := &wag.Config{
		Text:            textBuf,
		MemoryAlignment: os.Getpagesize(),
		Entry:           entry,
	}
	obj, err := wag.Compile(config, progReader, resolver{})
	if dumpText && len(obj.Text) > 0 {
		e := dump.Text(os.Stdout, obj.Text, textAddr, obj.FuncAddrs, &obj.Names)
		if err == nil {
			err = e
		}
	}
	if err != nil {
		log.Fatal(err)
	}

	setImportVectorCurrentMemory(obj.InitialMemorySize)

	globalsMemory, err := makeMem(obj.MemoryOffset+linearMemoryAddressSpace, syscall.PROT_NONE, 0)
	if err != nil {
		log.Fatal(err)
	}

	err = syscall.Mprotect(globalsMemory[:obj.MemoryOffset+obj.InitialMemorySize], syscall.PROT_READ|syscall.PROT_WRITE)
	if err != nil {
		log.Fatal(err)
	}

	copy(globalsMemory, obj.GlobalsMemory)

	memoryAddr := memAddr(globalsMemory) + uintptr(obj.MemoryOffset)

	if err := syscall.Mprotect(vecMem, syscall.PROT_READ); err != nil {
		log.Fatal(err)
	}

	if err := syscall.Mprotect(textMem, syscall.PROT_READ|syscall.PROT_EXEC); err != nil {
		log.Fatal(err)
	}

	stackMem, err := makeMem(stackSize, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_STACK)
	if err != nil {
		log.Fatal(err)
	}
	stackOffset := stackSize - len(obj.StackFrame)
	copy(stackMem[stackOffset:], obj.StackFrame)

	stackAddr := memAddr(stackMem)
	stackLimit := stackAddr + signalStackReserve
	stackPtr := stackAddr + uintptr(stackOffset)

	if stackLimit >= stackPtr {
		log.Fatal("stack is too small for starting program")
	}

	retaddr, retsize := exec(textAddr, stackLimit, memoryAddr, stackPtr)

	gasLeft := binary.LittleEndian.Uint64(contractData[8:])
	if gasLeft == 0xffffffffffffffff {
		fmt.Println("Out of gas")
	} else {
		fmt.Println("gas left: ", gasLeft, "result: ", globalsMemory[retaddr+obj.MemoryOffset:retaddr+obj.MemoryOffset+retsize])
	}
}
