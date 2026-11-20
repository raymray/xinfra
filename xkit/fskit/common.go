package fskit

import (
	"os"
	"path/filepath"

	"fmt"
)

// ensureDir 确保文件所在目录存在
func ensureDir(fPath string) error {
	dir := filepath.Dir(fPath)
	if dir == "" || dir == "." {
		return nil
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	return nil
}

// ensureExt 确保文件路径有指定后缀，如 ".csv"、".xlsx"、".json"
func ensureExt(fPath, ext string) string {
	if filepath.Ext(fPath) != ext {
		return fPath + ext
	}
	return fPath
}
