package infrastructure

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/otiai10/gosseract/v2"
)

func ExtractImagesFromPDF(pdfPath string) ([]string, error) {
	outputDir := "./output_images"
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("error creating diretory: %w", err)
	}

	cmd := exec.Command("pdfimages", "-j", pdfPath, outputDir+"/image")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("error executing pdfimages: %w", err)
	}

	files, err := ioutil.ReadDir(outputDir)
	if err != nil {
		return nil, fmt.Errorf("error reading directory images: %w", err)
	}

	var imagePaths []string
	for _, file := range files {
		if !file.IsDir() {
			imagePaths = append(imagePaths, outputDir+"/"+file.Name())
		}
	}

	return imagePaths, nil
}

// Função para realizar OCR em uma imagem
func PerformOCR(imagePath string) (string, error) {
	client := gosseract.NewClient()
	client.SetLanguage("por")
	defer client.Close()

	client.SetImage(imagePath)
	text, err := client.Text()
	if err != nil {
		log.Printf("error performing OCR: %v", err)
		return "", err
	}

	return removeEmptyLines(text), nil
}

// Função para remover linhas vazias
func removeEmptyLines(text string) string {
	lines := strings.Split(text, "\n")
	var nonEmptyLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}
	return strings.Join(nonEmptyLines, "\n")
}
