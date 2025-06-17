//go:build windows

package file

import (
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/go-sw/winfs/w32api"
)

var (
	// global progress callback for file operation
	progressCb uintptr
)

// initCb initializes a global callback to reduce callback limit pressure,
// this will only be called if callback is defined
func initCb() {
	if progressCb != 0 {
		return
	}

	// "runtime/syscall_windows.go".compileCallback has mutex by itself
	// and does the deduplication
	progressCb = w32api.NewCopyProgressRoutine(func(
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
		if data == nil {
			return w32api.PROGRESS_CONTINUE
		}
		handler := (*callbackHandler)(data).handler

		return handler.HandleRoutine(
			totalFileSize,
			totalBytesTransferred,
			streamSize,
			streamBytesTransferred,
			streamNumber,
			callbackReason,
			sourceFile,
			destinationFile,
		)
	})
}

// WinFile handles Windows specific file operation with progress
// callback and cancellation.
type WinFile struct {
	path     string // absoulte path with long path support
	callback callbackHandler
}

// NewWinFile returns new [WinFile] to be used for file operation,
// the specified file should exist.
func NewWinFile(path string) (*WinFile, error) {
	var err error
	var absPath string

	if len(path) >= 4 && (path[:4] == `\\?\` || path[:4] == `\??\`) {
		// path with prefix should be a absolute path
		absPath = path
	} else {
		absPath, err = filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		// prepend long path support prefix
		absPath = `\\?\` + absPath
	}

	// check if the specified file exists
	_, err = os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	return &WinFile{
		path: absPath,
	}, nil
}

// SetHandler sets handler which implements [progress].
func (f *WinFile) SetHandler(handler progress) {
	if handler == nil {
		return
	}

	initCb()

	f.callback.handler = handler
}

// Copy copies the underlying file to the destination, cancel pointer can optionally be a
// pointer to int32 value which can be set to non-zero value to cancel copy operation.
func (f *WinFile) Copy(destination string, options *CopyOptions, cancel *int32) error {
	return w32api.CopyFileEx(f.path, destination, progressCb, unsafe.Pointer(&f.callback), cancel, options.asFlags())
}

// Move moves the underlying file or directory to the destination.
func (f *WinFile) Move(destination string, options *MoveOptions) error {
	return w32api.MoveFileWithProgress(f.path, destination, progressCb, unsafe.Pointer(&f.callback), options.asFlags())
}

// GetShortName returns short path name(8.3 filename) of this file.
func (f *WinFile) GetShortName() (string, error) {
	return w32api.GetShortPathName(f.path)
}

// SetShortName sets short path name(8.3 filename) of this file, the
// file needs to be in NTFS volume.
func (f *WinFile) SetShortName(shortName string) error {
	u16Path, err := windows.UTF16PtrFromString(f.path)
	if err != nil {
		return err
	}

	var hnd windows.Handle
	defer func() {
		if err != nil {
			windows.CloseHandle(hnd)
		}
	}()

	hnd, err = windows.CreateFile(
		u16Path,
		windows.GENERIC_WRITE|windows.DELETE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT,
		0,
	)
	if err != nil {
		return err
	}

	if err = w32api.SetFileShortName(hnd, shortName); err != nil {
		return err
	}

	return windows.CloseHandle(hnd)
}
