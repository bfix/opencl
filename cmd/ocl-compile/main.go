package main

import (
	_ "embed"
	"log"
	"time"

	cl "github.com/CyberChainXyz/go-opencl"
)

//go:embed test.cl
var testSrc string

func main() {

	info, err := cl.Info()
	if err != nil {
		log.Println("no opencl device(s)")
		log.Fatal(err)
	}

	var device *cl.OpenCLDevice
loop:
	for _, pf := range info.Platforms {
		if pf.Name == "NVIDIA CUDA" {
			for _, dev := range pf.Devices {
				device = dev
				break loop
			}
		}
	}

	runner, err := device.InitRunner()
	if err != nil {
		log.Println("no runner")
		log.Fatal(err)
	}
	defer runner.Free()

	err = runner.CompileKernels([]string{testSrc}, []string{"test"}, "")
	if err != nil {
		log.Println("compile kernel failed")
		log.Fatal(err)
	}

	numbers := make([]int32, 1024)
	for i := 0; i < 1024; i++ {
		numbers[i] = int32(i)
	}
	result := make([]int32, 1024)

	inBuf, err := runner.CreateEmptyBuffer(cl.READ_ONLY, 4*len(numbers))
	if err != nil {
		log.Println("create input buffer failed")
		log.Fatal(err)
	}

	outBuf, err := runner.CreateEmptyBuffer(cl.WRITE_ONLY, 4*len(numbers))
	if err != nil {
		log.Println("create output buffer failed")
		log.Fatal(err)
	}

	args := []cl.KernelParam{
		cl.BufferParam(inBuf),
		cl.BufferParam(outBuf),
	}

	start := time.Now()
	for range 100000 {
		if err = cl.WriteBuffer(runner, 0, inBuf, numbers, true); err != nil {
			log.Println("writing input buffer failed")
			log.Fatal(err)
		}

		if err = runner.RunKernel(
			"test",
			1,
			[]uint64{0},
			[]uint64{1024},
			[]uint64{1024},
			args,
			true,
		); err != nil {
			log.Println("running kernel failed")
			log.Fatal(err)
		}

		if err = cl.ReadBuffer(runner, 0, outBuf, result); err != nil {
			log.Println("reading output buffer failed")
			log.Fatal(err)
		}
	}
	log.Printf("GPU Elapsed: %s\n", time.Since(start))

	start = time.Now()
	for range 100000 {
		for i, v := range numbers {
			result[i] = v * v
		}
	}
	log.Printf("CPU Elapsed: %s\n", time.Since(start))
}
