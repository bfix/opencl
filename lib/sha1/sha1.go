package sha1

import (
	_ "embed"
)

var Name = "sha1_kernel"

//go:embed sha1.cl
var Src string

type sha1Ocl struct {
}

func New() *sha1Ocl {
	return nil
}
