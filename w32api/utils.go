//go:build windows

package w32api

import (
	"bytes"
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

/*
	typedef struct _FILE_RENAME_INFO {
	#if _WIN32_WINNT >= _WIN32_WINNT_WIN10_RS1
		__C89_NAMELESS union {
			BOOLEAN ReplaceIfExists;
			DWORD Flags;
		};
	#else
		BOOLEAN ReplaceIfExists;
	#endif
		HANDLE RootDirectory;
		DWORD FileNameLength;
		WCHAR FileName[1];
	} FILE_RENAME_INFO,*PFILE_RENAME_INFO;
*/

// NewFileRenameInfo returns FILE_RENAME_INFO from winbase.h as bytes buffer used for renaming ADS name.
func NewFileRenameInfo(newName string, replace bool) ([]byte, error) {
	if len(newName) == 0 {
		return nil, errors.New("new name is empty")
	}

	alignSize := unsafe.Sizeof(windows.Handle(0)) // padding for alignment to match C struct

	var renameInfo bytes.Buffer
	var replaceIfExists uint32

	if replace {
		replaceIfExists = 1 // TRUE
	}

	renameInfo.Write(unsafe.Slice((*byte)(unsafe.Pointer(&replaceIfExists)), unsafe.Sizeof(replaceIfExists))) // set ReplaceIfExists DWORD(uint32) to true
	renameInfo.Write(make([]byte, alignSize-unsafe.Sizeof(replaceIfExists))) // alignment

	rootDirectory := windows.Handle(0)
	renameInfo.Write(unsafe.Slice((*byte)(unsafe.Pointer(&rootDirectory)), unsafe.Sizeof(rootDirectory))) // RootDirectory HANDLE(NULL)

	u16Name, err := windows.UTF16FromString(newName)
	if err != nil {
		return nil, err
	}

	// max length = 257(1(":") + 255(max ads name length) + 1(NULL termination))
	if len(u16Name) > 257 {
		return nil, errors.New("the length of new name exceeds max length(255)")
	}

	var fileNameLength uint32 = uint32((len(u16Name) - 1) * 2) // length in bytes without NULL terminaton

	renameInfo.Write(unsafe.Slice((*byte)(unsafe.Pointer(&fileNameLength)), unsafe.Sizeof(fileNameLength))) // FileNameLength DWORD(uint32)
	renameInfo.Write(unsafe.Slice((*byte)(unsafe.Pointer(&u16Name[0])), len(u16Name)*2)) // FileName []uint16(WCHAR[])

	return renameInfo.Bytes(), nil
}
