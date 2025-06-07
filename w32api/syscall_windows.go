//go:build windows

package w32api

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

// file operation functions

//sys	ntOpenFile(fileHandle *windows.Handle, accessMask uint32, objectAttributes *windows.OBJECT_ATTRIBUTES, ioStatusBlock *windows.IO_STATUS_BLOCK, sharedAccess uint32, openOptions uint32) (ntstatus error) = ntdll.NtOpenFile
//sys	ntClose(fileHandle windows.Handle) (ntstatus error) = ntdll.NtClose
//sys	ntQueryInformationFile(fileHandle windows.Handle, ioStatusBlock *windows.IO_STATUS_BLOCK, fileInformation unsafe.Pointer, length uint32, fileInformationClass int32) (ntstatus error) = ntdll.NtQueryInformationFile

func NtOpenFile(accessMask uint32, objectAttributes *windows.OBJECT_ATTRIBUTES, ioStatusBlock *windows.IO_STATUS_BLOCK, sharedAccess uint32, openOptions uint32) (fileHandle windows.Handle, err error) {
	err = ntOpenFile(&fileHandle, accessMask, objectAttributes, ioStatusBlock, sharedAccess, openOptions)
	return
}

func NtClose(fileHandle windows.Handle) (err error) {
	err = ntClose(fileHandle)
	return
}

func NtQueryInformationFile(fileHandle windows.Handle, ioStatusBlock *windows.IO_STATUS_BLOCK, fileInformation unsafe.Pointer, length uint32, fileInformationClass int32) (err error) {
	err = ntQueryInformationFile(fileHandle, ioStatusBlock, fileInformation, length, fileInformationClass)
	return
}

// ADS(alternate data stream) functions

//sys	findFirstStream(fileName *uint16, infoLevel int32, findStreamData unsafe.Pointer, flags uint32) (hnd windows.Handle, err error) [failretval==windows.InvalidHandle] = kernel32.FindFirstStreamW
//sys	findNextStream(findStream windows.Handle, findStreamData unsafe.Pointer) (err error) = kernel32.FindNextStreamW
//sys	findClose(findFile windows.Handle) (err error) = kernel32.FindClose

func FindFirstStream(fileName string, infoLevel int32, flags uint32) (hnd windows.Handle, data WIN32_FIND_STREAM_DATA, err error) {
	wStr, err := windows.UTF16PtrFromString(fileName)
	if err != nil {
		return
	}

	// returned error
	// windows.ERROR_HANDLE_EOF if there is no stream
	// windows.ERROR_INVALID_PARAMETER for unsupported file system
	hnd, err = findFirstStream(wStr, FindStreamInfoStandard, unsafe.Pointer(&data), flags) // flags should be 0
	return
}

func FindNextStream(findStream windows.Handle) (data WIN32_FIND_STREAM_DATA, err error) {
	err = findNextStream(findStream, unsafe.Pointer(&data))

	// returns windows.ERROR_HANDLE_EOF if there is no more stream
	return
}

func FindClose(findFile windows.Handle) (err error) {
	err = findClose(findFile)
	return
}

// backup functions

//sys	backupRead(file windows.Handle, buffer *byte, numberOfBytesToRead uint32, numberOfBytesRead *uint32, abort bool, processSecurity bool, context *uintptr) (err error) = kernel32.BackupRead
//sys	backupSeek(file windows.Handle, lowBytesToSeek uint32, highBytesToSeek uint32, lowBytesSeeked *uint32, highBytesSeeked *uint32, context *uintptr) (err error) = kernel32.BackupSeek
//sys	backupWrite(file windows.Handle, buffer *byte, numberOfBytesToRead uint32, numberOfBytesRead *uint32, abort bool, processSecurity bool, context *uintptr) (err error) = kernel32.BackupWrite

// BackupRead reads data associated with a specified file or directory.
//
// https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-backupread
func BackupRead(file windows.Handle, buffer []byte, abort, processSecurity bool, context *uintptr) (bytesRead uint32, err error) {
	err = backupRead(file, &buffer[0], uint32(len(buffer)), &bytesRead, abort, processSecurity, context)
	return
}

/*
BackupSeek skips portion of a data stream.

https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-backupseek

caveats:

If the current stream offset is in the middle of the data, it succeeds if
and only if the remaining data is not smaller than the seeked size, setting
the seeked size to the given size.

However, if the current stream offset is in the middle of backup header, it fails
without setting the last error value(eventually returns ERROR_SUCCESS(0x0))
and does nothing(offset does not increase).

If the given seek size is larger than the currently remaining data size, it fails
and sets the last error to ERROR_SEEK(0x19), but the offset actually advances to
the starting offset of next stream header while keeping the seeked size to 0.
*/
func BackupSeek(file windows.Handle, offset uint64, context *uintptr) (seeked uint64, err error) {
	var seekedLow, seekedHigh uint32
	err = backupSeek(file, uint32(offset), uint32(offset>>32), &seekedLow, &seekedHigh, context)
	return (uint64(seekedHigh) << 32) | uint64(seekedLow), err
}

// BackupWrite writes the associated data to specified file or directory.
//
// https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-backupwrite
func BackupWrite(file windows.Handle, buffer []byte, abort, processSecurity bool, context *uintptr) (bytesWritten uint32, err error) {
	err = backupWrite(file, &buffer[0], uint32(len(buffer)), &bytesWritten, abort, processSecurity, context)
	return
}

// EA(extended attributes) functions

//sys	ntSetEaFile(fileHandle windows.Handle, ioStatusBlock *windows.IO_STATUS_BLOCK, buffer unsafe.Pointer, length uint32) (ntstatus error) = ntdll.NtSetEaFile
//sys	ntQueryEaFile(fileHandle windows.Handle, ioStatusBlock *windows.IO_STATUS_BLOCK, buffer unsafe.Pointer, length uint32, returnSingleEntry bool, eaList unsafe.Pointer, eaListLength uint32, eaIndex *uint32, restartScan bool) (ntstatus error) = ntdll.NtQueryEaFile

func NtSetEaFile(fileHandle windows.Handle, ioStatusBlock *windows.IO_STATUS_BLOCK, buffer unsafe.Pointer, length uint32) (err error) {
	err = ntSetEaFile(fileHandle, ioStatusBlock, buffer, length)
	return
}

func NtQueryEaFile(fileHandle windows.Handle, ioStatusBlock *windows.IO_STATUS_BLOCK, buffer unsafe.Pointer, length uint32, returnSingleEntry bool, eaList unsafe.Pointer, eaListLength uint32, eaIndex *uint32, restartScan bool) (err error) {
	err = ntQueryEaFile(fileHandle, ioStatusBlock, buffer, length, returnSingleEntry, eaList, eaListLength, eaIndex, restartScan)
	return
}

// EFS(encrypted file system) functions


