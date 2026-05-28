package types

// FileOp 表示文件系统操作的类型。
type FileOp int

const (
	OpCreate FileOp = iota
	OpWrite
	OpRemove
	OpRename
	OpChmod
)

// FileEvent 是由 WatcherComponent 产生的领域无关的文件系统事件。
type FileEvent struct {
	Path string
	Op   FileOp
}
