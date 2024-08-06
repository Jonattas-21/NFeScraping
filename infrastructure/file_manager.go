package infrastructure

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/xuri/excelize/v2"
)

func CreateOutputDir() (string, error) {
	// Pasta onde a captura de tela será salva
	outputDir := "./output_screenshots"
	err := createDir(outputDir)
	if err != nil {
		return "", err
	}
	return outputDir, nil
}

func CreateExcelOutputDir() (string, error) {
	// Pasta onde o excel será salvo
	outputDir := "./output_sheets"
	err := createDir(outputDir)
	if err != nil {
		return "", err
	}
	return outputDir, nil
}

func createDir(outputDir string) error {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err = os.Mkdir(outputDir, os.ModePerm)
		if err != nil {
			log.Fatalf("error creating directory: %v", err)
			return err
		}
	}
	return nil
}

func CleanUp() {
	os.RemoveAll("./output_screenshots")
	os.RemoveAll("./output_images")
	//os.RemoveAll("./output_sheets/")
}

func CreateExcelOutput(columns []string, values [][]string) error {

	f := excelize.NewFile()
	sheetName := fmt.Sprintf("NFe_%d-%d", time.Now().Hour(), time.Now().Minute())
	index, err := f.NewSheet(sheetName)

	if err != nil {
		log.Fatalf("error creating output sheet result: %v", err)
		return err
	}

	for col, column := range columns {
		f.SetCellValue(sheetName, fmt.Sprintf("%s1", string(rune(65+col))), column)
	}

	for row, item := range values {
		row = row + 2

		for col, value := range item {
			f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune(65+col)), row), value)
		}
	}

	f.SetActiveSheet(index)
	if err := f.SaveAs(fmt.Sprintf("./output_sheets/%s.xlsx", sheetName)); err != nil {
		return err
	}

	return nil
}

func ExtractPageFromFileImageName(imageName string) (int, error) {
	// Exemplo de nome de arquivo: output_images/image-0001.ppm
	// O número da página é o último número do nome do arquivo
	var page int
	_, err := fmt.Sscanf(imageName, "output_images/image-%d.ppm", &page)
	if err != nil {
		return 0, err
	}

	return page, nil
}
