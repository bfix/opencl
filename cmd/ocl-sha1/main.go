package main

import (
	"crypto/sha1"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"log"
	"time"

	cl "github.com/CyberChainXyz/go-opencl"
)

var sha1Name = "sha1_kernel"

//go:embed sha1.cl
var sha1Src string

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

	err = runner.CompileKernels([]string{sha1Src}, []string{sha1Name}, "")
	if err != nil {
		log.Println("compile kernel failed")
		log.Fatal(err)
	}

	inBuf, err := runner.CreateEmptyBuffer(cl.READ_ONLY, 1028)
	if err != nil {
		log.Println("create input buffer failed")
		log.Fatal(err)
	}

	outBuf, err := runner.CreateEmptyBuffer(cl.WRITE_ONLY, 20)
	if err != nil {
		log.Println("create output buffer failed")
		log.Fatal(err)
	}

	in := make([]byte, 1028)
	out := make([]byte, 20)

	var size uint32 = 4
	binary.LittleEndian.PutUint32(in, size)
	copy(in[4:], []byte("test"))

	args := []cl.KernelParam{
		cl.BufferParam(inBuf),
		cl.BufferParam(outBuf),
	}

	start := time.Now()
	for range 10000 {
		if err = cl.WriteBuffer(runner, 0, inBuf, in, true); err != nil {
			log.Println("writing input buffer failed")
			log.Fatal(err)
		}

		if err = runner.RunKernel(
			sha1Name,
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

		if err = cl.ReadBuffer(runner, 0, outBuf, out); err != nil {
			log.Println("reading output buffer failed")
			log.Fatal(err)
		}

		for i := range 5 {
			v := binary.LittleEndian.Uint32(out[4*i : 4*i+4])
			binary.BigEndian.PutUint32(out[4*i:4*i+4], v)
		}
	}
	log.Printf("hash = %s\n", hex.EncodeToString(out))
	log.Printf("GPU Elapsed: %s\n", time.Since(start))

	start = time.Now()
	for range 10000 {
		hsh := sha1.New()
		hsh.Write(in[4:8])
		out = hsh.Sum(nil)
	}
	log.Printf("hash = %s\n", hex.EncodeToString(out))
	log.Printf("CPU Elapsed: %s\n", time.Since(start))
}
