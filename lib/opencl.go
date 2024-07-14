package lib

import (
	"errors"
	"fmt"

	cl "github.com/CyberChainXyz/go-opencl"
)

type OpenCLContext struct {
	info    *cl.OpenCLInfo
	devices map[string]*cl.OpenCLDevice
}

func _join(msg string, err error) error {
	return errors.Join(errors.New(msg), err)
}

func NewOpenCLContext() (ctx *OpenCLContext, err error) {
	ctx = &OpenCLContext{}
	if ctx.info, err = cl.Info(); err != nil {
		err = _join("no opencl device(s)", err)
		return
	}
	ctx.devices = make(map[string]*cl.OpenCLDevice)
	for _, pf := range ctx.info.Platforms {
		for _, dev := range pf.Devices {
			ctx.devices[dev.Name] = dev
		}
	}
	return
}

func (ctx *OpenCLContext) ListDevices() (list []string) {
	for dev := range ctx.devices {
		list = append(list, dev)
	}
	return
}

func (ctx *OpenCLContext) Prepare(devName string, src, label []string, num, inSize, outSize int) (runner *OpenCLRunner, err error) {
	dev, ok := ctx.devices[devName]
	if !ok {
		err = fmt.Errorf("unknown device '%s'", devName)
		return
	}
	runner = &OpenCLRunner{}
	runner.num = num
	if runner.inst, err = dev.InitRunner(); err != nil {
		err = _join("no runner", err)
		return
	}

	err = runner.inst.CompileKernels(src, label, "")
	if err != nil {
		err = _join("compile kernel failed", err)
		return
	}

	if runner.inBuf, err = runner.inst.CreateEmptyBuffer(cl.READ_ONLY, inSize); err != nil {
		err = _join("create input buffer failed", err)
		return
	}

	if runner.outBuf, err = runner.inst.CreateEmptyBuffer(cl.WRITE_ONLY, outSize); err != nil {
		err = _join("create output buffer failed", err)
		return
	}

	runner.args = []cl.KernelParam{
		cl.BufferParam(runner.inBuf),
		cl.BufferParam(runner.outBuf),
	}
	return
}

type OpenCLRunner struct {
	inst   *cl.OpenCLRunner
	inBuf  *cl.Buffer
	outBuf *cl.Buffer
	args   []cl.KernelParam
	num    int
}

func CopyIn[T any](r *OpenCLRunner, offset int, in []T) (err error) {
	return cl.WriteBuffer(r.inst, offset, r.inBuf, in, true)
}

func (r *OpenCLRunner) Run(name string) (err error) {
	return r.inst.RunKernel(
		name,
		1,
		nil,
		[]uint64{uint64(r.num)},
		[]uint64{uint64(r.num)},
		r.args,
		true,
	)
}

func CopyOut[T any](r *OpenCLRunner, offset int, out []T) (err error) {
	return cl.ReadBuffer(r.inst, offset, r.outBuf, out)
}
