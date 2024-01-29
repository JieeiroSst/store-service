package excel

import (
	"io"
	"mime/multipart"
	"os"

	"github.com/tealeg/xlsx"
)

func ReadFileExcel(f *multipart.FileHeader) (*xlsx.File, error) {
	tempFile, err := os.CreateTemp("", "excel-*.xlsx")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempFile.Name())

	src, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	_, err = io.Copy(tempFile, src)
	if err != nil {
		return nil, err
	}
	xlsxFile, err := xlsx.OpenFile(tempFile.Name())
	if err != nil {
		return nil, err
	}

	return xlsxFile, nil
}

func GetRowValues(sheet *xlsx.Sheet, row int) []map[string]interface{} {
	rows := sheet.MaxRow
	cols := sheet.MaxCol

	data := make([]map[string]interface{}, 0)

	for r := 0; r <= rows; r++ {
		row := make(map[string]interface{})
		for c := 0; c <= cols; c++ {
			cell := sheet.Cell(r, c)
			column := sheet.Row(r).Cells[c].String()

			types := cell.Type()

			switch types {
			case xlsx.CellTypeString:
				valueStr := cell.String()
				row[column] = valueStr
			case xlsx.CellTypeNumeric:
				valueNum, _ := cell.Float()
				row[column] = valueNum
			case xlsx.CellTypeBool:
				valueBool := cell.Bool()
				row[column] = valueBool
			}
		}
		data = append(data, row)
	}

	return data
}
