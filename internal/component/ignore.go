package component

import (
	"path/filepath"
	"strings"
)

// IgnoreComponent 管理基于监控根目录的忽略路径规则。
type IgnoreComponent struct {
	rootAbs string
	ignored map[string]struct{}
}

// NewIgnoreComponent 创建 IgnoreComponent。
func NewIgnoreComponent(root string, ignored []string) *IgnoreComponent {
	rootPath := filepath.Clean(filepath.FromSlash(root))
	rootAbs, err := filepath.Abs(rootPath)
	if err != nil {
		rootAbs = rootPath
	}

	component := &IgnoreComponent{
		rootAbs: rootAbs,
		ignored: make(map[string]struct{}),
	}
	for _, path := range ignored {
		component.add(path)
	}
	return component
}

// ShouldIgnore 报告路径是否命中忽略规则。
func (ignore *IgnoreComponent) ShouldIgnore(path string) bool {
	if ignore == nil || len(ignore.ignored) == 0 {
		return false
	}

	relPath := ignore.relativePath(path)
	if _, ok := ignore.ignored[relPath]; ok {
		return true
	}
	for ignoredPath := range ignore.ignored {
		if hasPathPrefix(relPath, ignoredPath) {
			return true
		}
	}
	return false
}

func (ignore *IgnoreComponent) add(path string) {
	relPath := cleanRelativePath(path)
	if relPath == "" || relPath == "." || isParentPath(relPath) {
		return
	}
	ignore.ignored[relPath] = struct{}{}
}

func (ignore *IgnoreComponent) relativePath(path string) string {
	cleanPath := filepath.Clean(filepath.FromSlash(path))
	if filepath.IsAbs(cleanPath) {
		return relFromRoot(ignore.rootAbs, cleanPath)
	}

	absPath, err := filepath.Abs(cleanPath)
	if err == nil {
		if relPath := relFromRoot(ignore.rootAbs, absPath); !isParentPath(relPath) {
			return relPath
		}
	}
	return cleanRelativePath(cleanPath)
}

func relFromRoot(root string, path string) string {
	relPath, err := filepath.Rel(root, path)
	if err != nil {
		return cleanRelativePath(path)
	}
	return cleanRelativePath(relPath)
}

func cleanRelativePath(path string) string {
	return filepath.Clean(filepath.FromSlash(path))
}

func hasPathPrefix(path string, prefix string) bool {
	if path == prefix {
		return true
	}
	separator := string(filepath.Separator)
	return strings.HasPrefix(path, prefix+separator)
}

func isParentPath(path string) bool {
	return path == ".." || strings.HasPrefix(path, ".."+string(filepath.Separator))
}
