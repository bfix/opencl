package main

import (
	gsha1 "crypto/sha1"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"log"
	"opencl/lib"
	"opencl/lib/sha1"
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

	srcs := []string{sha1.Src}
	labels := []string{sha1.Name}
	num := 1000
	in := make([]byte, 68*num)
	out := make([]int32, 5*num)

	for i := range num {
		binary.LittleEndian.PutUint32(in[i*68:], 4)
		copy(in[i*68+4:], []byte("test"))
	}

	runner, err := ctx.Prepare(devName, srcs, labels, num, num*68, num*20)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	for range 10 {
		if err = lib.CopyIn(runner, 0, in); err != nil {
			log.Println("writing input buffer failed")
			log.Fatal(err)
		}
		if err = runner.Run(sha1.Name); err != nil {
			log.Println("running kernel failed")
			log.Fatal(err)
		}

		if err = lib.CopyOut(runner, 0, out); err != nil {
			log.Println("reading output buffer failed")
			log.Fatal(err)
		}
	}
	log.Printf("GPU Elapsed: %s\n", time.Since(start))
	log.Printf("hash = %08x%08x%08x%08x%08x\n",
		uint32(out[0]), uint32(out[1]), uint32(out[2]),
		uint32(out[3]), uint32(out[4]))
	log.Println()

	out2 := make([]byte, 20*num)
	start = time.Now()
	for range 10000 {
		hsh := gsha1.New()
		hsh.Write(in[4:8])
		out2 = hsh.Sum(nil)
	}
	log.Printf("CPU Elapsed: %s\n", time.Since(start))
	log.Printf("hash = %s\n", hex.EncodeToString(out2))
}
