package main

import (
	_ "embed"
	"flag"
	"log"
	"opencl/lib"
	"opencl/lib/test"
	"time"
)

func main() {

	var devName string
	flag.StringVar(&devName, "d", "", "device name")
	flag.Parse()

	ctx, err := lib.NewOpenCLContext()
	if err != nil {
		log.Fatal(err)
	}
	if len(devName) == 0 {
		for i, dev := range ctx.ListDevices() {
			log.Printf("%d: %s\n", i+1, dev)
		}
		return
	}

	srcs := []string{test.Src}
	labels := []string{test.Name}
	num := 1024
	in := make([]int32, num)
	out := make([]int32, num)

	for i := range num {
		in[i] = int32(i)
	}

	runner, err := ctx.Prepare(devName, srcs, labels, num, num*4, num*4)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	for range 100000 {
		if err = lib.CopyIn(runner, 0, in); err != nil {
			log.Println("writing input buffer failed")
			log.Fatal(err)
		}

		if err = runner.Run(test.Name); err != nil {
			log.Println("running kernel failed")
			log.Fatal(err)
		}

		if err = lib.CopyOut(runner, 0, out); err != nil {
			log.Println("reading output buffer failed")
			log.Fatal(err)
		}
	}
	log.Printf("GPU Elapsed: %s\n", time.Since(start))

	start = time.Now()
	for range 100000 {
		for i, v := range in {
			out[i] = v * v
		}
	}
	log.Printf("CPU Elapsed: %s\n", time.Since(start))
}
