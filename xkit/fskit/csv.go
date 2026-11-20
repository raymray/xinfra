package fskit

import (
	"bytes"
	"encoding/csv"
	"os"

	"fmt"
	"github.com/spf13/cast"
)

// _utf8BOM UTF-8 BOM 标记
var _utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// ReadCSV 读取 CSV 文件全部行，自动去除 UTF-8 BOM
func ReadCSV(fPath string) ([][]string, error) {
	data, err := os.ReadFile(fPath)
	if err != nil {
		return nil, fmt.Errorf("read file, %s: %w", fPath, err)
	}
	// 去除 UTF-8 BOM
	data = bytes.TrimPrefix(data, _utf8BOM)

	r := csv.NewReader(bytes.NewReader(data))
	// 允许字段中包含换行符
	r.LazyQuotes = true
	r.FieldsPerRecord = -1

	return r.ReadAll()
}

// ReadCSVAsMaps 读取 CSV 文件为 []map[string]string，第一行作为 header
func ReadCSVAsMaps(fPath string) ([]string, []map[string]any, error) {
	data, err := os.ReadFile(fPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read file, %s: %w", fPath, err)
	}
	data = bytes.TrimPrefix(data, _utf8BOM)

	r := csv.NewReader(bytes.NewReader(data))
	r.LazyQuotes = true
	r.FieldsPerRecord = -1

	records, err := r.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("parse csv, %s: %w", fPath, err)
	}
	if len(records) == 0 {
		return nil, nil, nil
	}

	headers := records[0]
	result := make([]map[string]any, 0, len(records)-1)
	for _, row := range records[1:] {
		m := make(map[string]any, len(headers))
		for i, h := range headers {
			if i < len(row) {
				m[h] = row[i]
			}
		}
		result = append(result, m)
	}
	return headers, result, nil
}

// WriteCSV 将 map 数组写入 CSV 文件，自动添加 UTF-8 BOM
func WriteCSV(data []map[string]any, fPath string, headers ...string) (err error) {
	if len(data) == 0 {
		return nil
	}

	if len(headers) == 0 {
		for k := range data[0] {
			headers = append(headers, k)
		}
	}

	fPath = ensureExt(fPath, ".csv")
	if err = ensureDir(fPath); err != nil {
		return
	}

	f, err := os.Create(fPath)
	if err != nil {
		return fmt.Errorf("create csv file, %s: %w", fPath, err)
	}
	defer func() {
		if cerr := f.Close(); err == nil && cerr != nil {
			err = fmt.Errorf("close csv file, %s: %w", fPath, cerr)
		}
	}()

	// 写入 UTF-8 BOM
	if _, err = f.Write(_utf8BOM); err != nil {
		return fmt.Errorf("write bom: %w", err)
	}

	w := csv.NewWriter(f)

	if err = w.Write(headers); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	for _, row := range data {
		record := make([]string, 0, len(headers))
		for _, k := range headers {
			record = append(record, cast.ToString(row[k]))
		}
		if err = w.Write(record); err != nil {
			return fmt.Errorf("write row: %w", err)
		}
	}

	// 显式 Flush 并检查错误：缓冲区写盘失败（如磁盘满）不能被吞掉
	w.Flush()
	if err = w.Error(); err != nil {
		return fmt.Errorf("flush csv, %s: %w", fPath, err)
	}
	return nil
}
