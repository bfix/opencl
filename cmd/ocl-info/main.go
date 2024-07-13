package main

import (
	cl "github.com/CyberChainXyz/go-opencl"
	"github.com/kr/pretty"
)

func main() {
	info, err := cl.Info()
	if err != nil {
		panic("no opencl device")
	}
	pretty.Println(info)
}
