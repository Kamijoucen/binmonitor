package component

// FileOp represents the type of file system operation.
type FileOp int

const (
	OpCreate FileOp = iota
	OpWrite
	OpRemove
	OpRename
	OpChmod
)

// FileEvent is a domain-neutral file system event produced by WatcherComponent.
type FileEvent struct {
	Path string
	Op   FileOp
}
