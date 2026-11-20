package fskit

import (
	"sort"

	"fmt"
	"github.com/raymray/xinfra/xkit"
	"github.com/xuri/excelize/v2"
)

const _defaultSheetName = "Sheet1"

// ReadExcelColumns 流式读取 Excel 指定列数据
// 使用 Rows 迭代器逐行读取，避免全量加载到内存
func ReadExcelColumns(fPath, sheetName string, skipHeader bool, columnNames ...string) (data [][]string, err error) {
	var f *excelize.File
	if f, err = excelize.OpenFile(fPath); err != nil {
		return nil, fmt.Errorf("open file, %s: %w", fPath, err)
	}
	defer func() { _ = f.Close() }()

	if sheetName == "" {
		return nil, fmt.Errorf("sheet name is empty")
	}

	rows, err := f.Rows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("read sheet rows, %s: %w", sheetName, err)
	}
	defer func() { _ = rows.Close() }()

	// 读取第一行确定列索引
	if !rows.Next() {
		return nil, nil
	}
	headers, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("read header row, %s: %w", sheetName, err)
	}

	headersIdx := make(map[string]int, len(headers))
	for idx, col := range headers {
		headersIdx[col] = idx
	}

	// 按 columnNames 的顺序匹配列索引
	var colIndices []int
	var matchedHeaders []string
	for _, name := range columnNames {
		if idx, ok := headersIdx[name]; ok {
			colIndices = append(colIndices, idx)
			matchedHeaders = append(matchedHeaders, name)
		}
	}

	if !skipHeader {
		data = append(data, matchedHeaders)
	}

	// 逐行读取数据
	for rows.Next() {
		row, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("read data row, %s: %w", sheetName, err)
		}
		var item []string
		for _, idx := range colIndices {
			if idx < len(row) {
				item = append(item, row[idx])
			}
		}
		if len(item) > 0 {
			data = append(data, item)
		}
	}

	return data, rows.Error()
}

// ReadExcelSheet 流式读取 Excel 整行数据，通过 MChannel 管道推送
func ReadExcelSheet(fPath, sheetName string, pipe *xkit.MChannel[[]string]) (err error) {
	var f *excelize.File
	if f, err = excelize.OpenFile(fPath); err != nil {
		return fmt.Errorf("open file, %s: %w", fPath, err)
	}
	defer func() { _ = f.Close() }()

	if sheetName == "" {
		return fmt.Errorf("sheet name is empty")
	}

	rows, err := f.Rows(sheetName)
	if err != nil {
		return fmt.Errorf("read sheet rows, %s: %w", sheetName, err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		columns, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("read row columns, %s: %w", sheetName, err)
		}
		pipe.Write(columns)
	}
	pipe.End()
	return rows.Error()
}

// WriteExcel 流式将 map 数组写入 Excel 文件
// 使用 StreamWriter 逐行写入，减少内存占用，适合大数据量
func WriteExcel(data []map[string]any, fPath, sheetName string, headers ...string) (err error) {
	if len(data) == 0 {
		return nil
	}

	if len(headers) == 0 {
		for k := range data[0] {
			headers = append(headers, k)
		}
		sort.Strings(headers)
	}

	fPath = ensureExt(fPath, ".xlsx")
	if err = ensureDir(fPath); err != nil {
		return
	}

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	if sheetName == "" {
		sheetName = _defaultSheetName
	}
	if sheetName != _defaultSheetName {
		if err = f.SetSheetName(_defaultSheetName, sheetName); err != nil {
			return fmt.Errorf("rename sheet %s: %w", sheetName, err)
		}
	}

	sw, err := f.NewStreamWriter(sheetName)
	if err != nil {
		return fmt.Errorf("create stream writer, %s: %w", sheetName, err)
	}

	// 写表头
	headerCells := make([]any, len(headers))
	for i, h := range headers {
		headerCells[i] = h
	}
	if err = sw.SetRow("A1", headerCells); err != nil {
		_ = sw.Flush()
		return fmt.Errorf("write header row: %w", err)
	}

	// 流式写数据行
	for rowIdx, row := range data {
		cell, _ := excelize.CoordinatesToCellName(1, rowIdx+2)
		cells := make([]any, len(headers))
		for colIdx, h := range headers {
			if v, ok := row[h]; ok {
				cells[colIdx] = v
			} else {
				cells[colIdx] = ""
			}
		}
		if err = sw.SetRow(cell, cells); err != nil {
			_ = sw.Flush()
			return fmt.Errorf("write row %d: %w", rowIdx+2, err)
		}
	}

	if err = sw.Flush(); err != nil {
		return fmt.Errorf("flush stream writer: %w", err)
	}

	return f.SaveAs(fPath)
}
