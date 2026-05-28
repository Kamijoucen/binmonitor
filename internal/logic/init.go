package logic

import (
	"os"
	"path/filepath"

	"binmonitor/internal/appctx"
)

// InitStateFromPath 递归扫描 root 并记录所有文件大小到状态组件中。
func InitStateFromPath(appCtx *appctx.AppCtx, root string) error {
	state := appCtx.State()
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// 权限错误或其他遍历错误：跳过但继续
			return nil
		}
		if appCtx.Ignore().ShouldIgnore(path) {
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
