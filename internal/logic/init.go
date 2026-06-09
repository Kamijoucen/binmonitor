package logic

import (
	"os"
	"path/filepath"
)

// FileSizeSetter 定义设置文件大小的能力。由 Component 实现。
type FileSizeSetter interface {
	SetSize(path string, size int64)
}

// PathIgnorer 定义判断路径是否应被忽略的能力。由 Component 实现。
type PathIgnorer interface {
	ShouldIgnore(path string) bool
}

// InitStateFromPath 递归扫描 root 并记录所有文件大小到状态组件中。
func InitStateFromPath(state FileSizeSetter, ignore PathIgnorer, root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// 权限错误或其他遍历错误：跳过但继续
			return nil
		}
		if ignore.ShouldIgnore(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.IsDir() {
			state.SetSize(path, info.Size())
		}
		return nil
	})
}
