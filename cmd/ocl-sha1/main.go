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
	inSize := 68 * num
	in := make([]byte, inSize)
	outSize := 20 * num
	out := make([]byte, outSize)

	for i := range num {
		binary.LittleEndian.PutUint32(in[i*68:], 4)
		copy(in[i*68+4:], []byte("test"))
	}

	runner, err := ctx.Prepare(devName, srcs, labels, num, inSize, outSize)
	if err != nil {
		log.Fatal(err)
	}

	start := time.Now()
	for range 10 {
		if err = runner.CopyIn(0, in); err != nil {
			log.Println("writing input buffer failed")
			log.Fatal(err)
		}
		if err = runner.Run(sha1.Name); err != nil {
			log.Println("running kernel failed")
			log.Fatal(err)
		}

		if err = runner.CopyOut(0, out); err != nil {
			log.Println("reading output buffer failed")
			log.Fatal(err)
		}

		for i := range 5 * num {
			v := binary.LittleEndian.Uint32(out[4*i : 4*i+4])
			binary.BigEndian.PutUint32(out[4*i:4*i+4], v)
		}
	}
	log.Printf("hash = %s\n", hex.EncodeToString(out[:20]))
	log.Printf("GPU Elapsed: %s\n", time.Since(start))

	start = time.Now()
	for range 10000 {
		hsh := gsha1.New()
		hsh.Write(in[4:8])
		out = hsh.Sum(nil)
	}
	log.Printf("hash = %s\n", hex.EncodeToString(out))
	log.Printf("CPU Elapsed: %s\n", time.Since(start))
}
