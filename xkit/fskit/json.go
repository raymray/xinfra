package fskit

import (
	"encoding/json"
	"os"

	"fmt"
)

// ReadJSON 泛型读取 JSON 文件
func ReadJSON[T any](fPath string) (result T, err error) {
	data, err := os.ReadFile(fPath)
	if err != nil {
		return result, fmt.Errorf("read file, %s: %w", fPath, err)
	}
	if err = json.Unmarshal(data, &result); err != nil {
		return result, fmt.Errorf("unmarshal json, %s: %w", fPath, err)
	}
	return result, nil
}

// WriteJSON 将数据写入 JSON 文件
// indents 可选参数控制是否格式化输出（默认 false）
func WriteJSON(data any, fPath string, indents ...bool) (err error) {
	if data == nil {
		return nil
	}

	fPath = ensureExt(fPath, ".json")
	if err = ensureDir(fPath); err != nil {
		return
	}

	var buf []byte
	indent := len(indents) > 0 && indents[0]
	if indent {
		buf, err = json.MarshalIndent(data, "", "  ")
	} else {
		buf, err = json.Marshal(data)
	}
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	if err = os.WriteFile(fPath, buf, 0644); err != nil {
		return fmt.Errorf("write file, %s: %w", fPath, err)
	}
	return nil
}
