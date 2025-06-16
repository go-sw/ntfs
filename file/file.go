//go:build windows

package file

import (
	"path/filepath"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/go-sw/ntfs/w32api"
)

var (
	progressCb uintptr
	cbMutex    sync.Mutex
)

// prevent creating multiple callbacks
func initCb() {
	cbMutex.Lock()
	defer cbMutex.Unlock()

	if progressCb != 0 {
		return
	}

	progressCb = windows.NewCallback(func(
		totalFileSize,
		totalBytesTransferred,
		streamSize,
		streamBytesTransferred int64,
		streamNumber,
		callbackReason uint32,
		sourceFile,
		destinationFile windows.Handle,
		data unsafe.Pointer,
	) uintptr {
		

	})
}

type Progress struct{

}

type ProgressHolder struct {
	
}

// WinFile handles Windows specific file operation with progress callback and cancellation.
type WinFile struct {
	path string
}
