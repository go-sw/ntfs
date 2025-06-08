package efs

import (
	"errors"
	"io"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/go-sw/ntfs/w32api"
)

type ExportContext struct {
	Context any

	// If the function succeeds, it must set the return value to ERROR_SUCCESS.
	Handler func(data *byte, ctx *ExportContext, length uint32) uintptr
}

type ImportContext struct {
	Context any

	// If the function succeeds, it must set the return value to ERROR_SUCCESS, and set the value pointed to by the length parameter to the number of bytes copied into data.
	//
	// When the end of the backup file is reached, set ulLength to zero to tell the system that the entire file has been processed.
	Handler func(data *byte, ctx *ImportContext, length *uint32) uintptr // user defined callback
}

// newCallback creates callback function pointer to be used by ReadEncryptedFileRaw or WriteEncryptedFileRaw.
// Note that only a limited number of callbacks may be created in a single Go process.
func newCallback[userCtx *ExportContext | *ImportContext](ctx userCtx) uintptr {
	switch any(ctx).(type) {
	case *ExportContext:
		cb := func(data *byte, callbackContext unsafe.Pointer, length uint32) uintptr {
			c := (*ExportContext)(callbackContext)

			return c.Handler(data, c, length)
		}
		return windows.NewCallback(cb)
	case *ImportContext:
		cb := func(data *byte, callbackContext unsafe.Pointer, length *uint32) uintptr {
			c := (*ImportContext)(callbackContext)

			return c.Handler(data, c, length)
		}
		return windows.NewCallback(cb)
	default:
		// cannot reach here
		return 0
	}
}

type EfsClient struct {
	readcb, writeCb uintptr

	ReadCtx  *ExportContext
	WriteCtx *ImportContext

	// TODO: handle certificate(wincrypt)
}

func NewEfsClient() (*EfsClient, error) {
	client := &EfsClient{
		ReadCtx:  &ExportContext{},
		WriteCtx: &ImportContext{},
	}

	client.readcb = newCallback(client.ReadCtx)
	client.writeCb = newCallback(client.WriteCtx)

	return client, nil
}

type rawReadWriteCtx struct {
	target io.ReadWriter
	err    error
}

// RawReadWriter converts encrypted file or directory into a raw data keeping its encryption.
type RawReadWriter struct {
	*EfsClient

	ctx *rawReadWriteCtx
}

func NewRawReadWriter() (*RawReadWriter, error) {
	client, err := NewEfsClient()
	if err != nil {
		return nil, err
	}

	rw := &RawReadWriter{
		EfsClient: client,
		ctx:       &rawReadWriteCtx{},
	}

	rw.ReadCtx.Context = rw.ctx
	rw.WriteCtx.Context = rw.ctx

	rw.ReadCtx.Handler = func(data *byte, ctx *ExportContext, length uint32) uintptr {
		c := ctx.Context.(*rawReadWriteCtx)

		buf := unsafe.Slice(data, length)
		_, err := c.target.Write(buf)
		if err != nil {
			c.err = err
			return uintptr(windows.ERROR_WRITE_FAULT)
		}

		return uintptr(windows.ERROR_SUCCESS)
	}
	rw.WriteCtx.Handler = func(data *byte, ctx *ImportContext, length *uint32) uintptr {
		c := ctx.Context.(*rawReadWriteCtx)

		buf := unsafe.Slice(data, *length)
		n, err := c.target.Read(buf)
		if err == io.EOF {
			*length = 0
		} else if err != nil {
			c.err = err
			return uintptr(windows.ERROR_READ_FAULT)
		} else {
			*length = uint32(n)
		}

		return uintptr(windows.ERROR_SUCCESS)
	}

	return rw, nil
}

// ReadRaw reads encrypted file as a raw stream then writes it to dst.
func (rw *RawReadWriter) ReadRaw(srcFile string, dst io.ReadWriter) error {
	stat, err := os.Stat(srcFile)
	if err != nil {
		return err
	}

	var openFlag uint32 = 0
	if stat.IsDir() {
		openFlag |= w32api.CREATE_FOR_DIR
	}

	rw.ctx.target = dst

	rawCtx, err := w32api.OpenEncryptedFileRaw(srcFile, openFlag)
	if err != nil {
		return err
	}
	defer w32api.CloseEncryptedFileRaw(rawCtx)

	err = w32api.ReadEncryptedFileRaw(rw.readcb, unsafe.Pointer(rw.ReadCtx), rawCtx)
	if err != nil {
		return errors.Join(err, rw.ctx.err)
	}

	return nil
}

// WriteRaw reads data from src then writes it as a encrypted file,
// if dir is true data is written as a directory with encrypted files.
func (rw *RawReadWriter) WriteRaw(dstFile string, src io.ReadWriter, dir bool) error {
	var openFlag uint32 = w32api.CREATE_FOR_IMPORT
	if dir {
		openFlag |= w32api.CREATE_FOR_DIR
	}

	rw.ctx.target = src

	rawCtx, err := w32api.OpenEncryptedFileRaw(dstFile, openFlag)
	if err != nil {
		return err
	}
	defer w32api.CloseEncryptedFileRaw(rawCtx)

	err = w32api.WriteEncryptedFileRaw(rw.writeCb, unsafe.Pointer(rw.WriteCtx), rawCtx)
	if err != nil {
		return errors.Join(err, rw.ctx.err)
	}

	return nil
}
