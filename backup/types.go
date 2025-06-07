package backup

// StreamType indicates the type of data in this backup stream.
//
// Can be used for handling data based on the type.
type StreamType uint32

const (
	BackupInvalid StreamType = iota
	BackupData
	BackupEaData
	BackupSecurityData
	BackupAlternateData
	BackupLink
	BackupPropertyData
	BackupObjectId
	BackupReparseData
	BackupSparseBlock
	BackupTxfsData
	BackupGhostedFileExtents
)

// Stream attributes indicates properties of the backup stream.
const (
	StreamNormalAttribute  uint32 = 0
	StreamModifiedWhenRead uint32 = 1 << (iota - 1)
	StreamContainsSecurity
	StreamContainsProperties
	StreamSparseAttribute
	StreamContainsGhostedFileExtents
)
