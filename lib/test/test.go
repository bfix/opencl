package test

import (
	_ "embed"
)

//go:embed test.cl
var Src string

var Name = "test"
