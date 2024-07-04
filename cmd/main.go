package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"go/format"
	"net/http"
	"os"
	"strings"
)

const (
	modelsFileURL = "https://raw.githubusercontent.com/KHwang9883/MobileModels-csv/main/models.csv"
	mobileType    = "mob"
	brandMapFile  = "brand_map.go"
)

// 从 github 获取手机型号和品牌信息后替换 brand_map.go 文件
func main() {
	resp, err := http.Get(modelsFileURL)
	if err != nil {
		fmt.Println("Error fetching the models file:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: received non-200 response code: %d\n", resp.StatusCode)
		return
	}

	reader := csv.NewReader(resp.Body)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading CSV data:", err)
		return
	}

	brands := make(map[string]string, len(records))
	models := make([]string, 0, len(records))
	for _, record := range records[1:] { // Skip header row
		if record[1] != mobileType { // Skip non-mobile records
			continue
		}

		model := strings.ToLower(record[0])
		brand := strings.ToLower(record[2])

		if _, ok := brands[model]; !ok {
			models = append(models, model)
		}

		brands[model] = normalizeBrand(brand)
	}

	var buf bytes.Buffer
	buf.WriteString("// Code generated. DO NOT EDIT.\n\n")
	buf.WriteString("//go:generate go run ./cmd\n")
	buf.WriteString("package useragentparser\n")
	buf.WriteString("var brandMap = map[string]string{\n")
	for _, model := range models { // Skip header row
		buf.WriteString("  \"" + model + "\": \"" + brands[model] + "\",\n")
	}
	buf.WriteString("}")

	fmtSrc, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Println("Error formatting source code:", err)
		return
	}

	err = os.WriteFile(brandMapFile, fmtSrc, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func normalizeBrand(brand string) string {
	switch brand {
	case "realme":
		return "oppo"
	case "zhixuan":
		return "huawei"
	default:
		return brand
	}
}
