package sha1

import (
	_ "embed"
)

var Name = "sha1_kernel"

//go:embed sha1.cl
var Src string

type WorkIn struct {
	size int32
	msg  [64]byte
}

func NewWorkIn(msg []byte) (w WorkIn) {
	w.size = int32(len(msg))
	copy(w.msg[:], msg)
	return
}

type WorkOut [5]uint32

type sha1Ocl struct {
}

func New() *sha1Ocl {
	return nil
}
