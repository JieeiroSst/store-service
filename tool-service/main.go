package main

import (
	"fmt"
	"log"
	"reflect"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gin-gonic/gin"
)

type Author struct {
	Name   string
	Gender string
	Age    int
	Status string
}

func main() {
	r := gin.Default()
	xlsx := excelize.NewFile()

	sheet1Name := "Sheet One"
	xlsx.SetSheetName(xlsx.GetSheetName(1), sheet1Name)

	xlsx.SetCellValue(sheet1Name, "A1", "Name")
	xlsx.SetCellValue(sheet1Name, "B1", "Gender")
	xlsx.SetCellValue(sheet1Name, "C1", "Age")
	xlsx.SetCellValue(sheet1Name, "D1", "Status")

	err := xlsx.AutoFilter(sheet1Name, "A1", "D1", "")
	if err != nil {
		log.Fatal("ERROR", err.Error())
	}

	structArray := []Author{
		{
			Name:   "Noval",
			Gender: "male",
			Age:    18,
			Status: "Thành Công",
		},
		{
			Name:   "Noval",
			Gender: "female",
			Age:    18,
			Status: "Thành Công",
		},
		{
			Name:   "Noval",
			Gender: "female",
			Age:    19,
			Status: "Huỷ",
		},
	}

	var data []map[string]interface{}
	for _, s := range structArray {
		data = append(data, structToMap(s))
	}

	for i, each := range data {
		xlsx.SetCellValue(sheet1Name, fmt.Sprintf("A%d", i+2), each["Name"])
		xlsx.SetCellValue(sheet1Name, fmt.Sprintf("B%d", i+2), each["Gender"])
		xlsx.SetCellValue(sheet1Name, fmt.Sprintf("C%d", i+2), each["Age"])
		xlsx.SetCellValue(sheet1Name, fmt.Sprintf("D%d", i+2), each["Status"])
	}
	r.GET("/", func(c *gin.Context) {
		buffer, _ := xlsx.WriteToBuffer()
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename=data.xlsx")
		c.Writer.Write(buffer.Bytes())
	})
	r.Run()
}

func structToMap(s interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	t := reflect.TypeOf(s)
	v := reflect.ValueOf(s)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i).Interface()
		m[field.Name] = value
	}
	return m
}
