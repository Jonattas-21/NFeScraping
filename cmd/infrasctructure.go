package main

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
		return nil, fmt.Errorf("erro ao criar diretório para imagens: %w", err)
	}

	cmd := exec.Command("pdfimages", "-j", pdfPath, outputDir+"/image")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("erro ao executar pdfimages: %w", err)
	}

	files, err := ioutil.ReadDir(outputDir)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler diretório de imagens: %w", err)
	}

	var imagePaths []string
	for _, file := range files {
		if !file.IsDir() {
			imagePaths = append(imagePaths, outputDir+"/"+file.Name())
		}
	}

	return imagePaths, nil
}

// Função para verificar se o texto contém padrões de uma nota fiscal
func IsNotaFiscal(text string) bool {

	palavrasChave := []string{"NOTA FISCAL", "NFS-e"}

	count := len(palavrasChave)
	for _, palavra := range palavrasChave {
		if strings.Contains(strings.ToUpper(text), strings.ToUpper(palavra)) {
			count--
		}
	}

	return count == 0
}

// Função para realizar OCR em uma imagem
func PerformOCR(imagePath string) (string, error) {
	client := gosseract.NewClient()
	client.SetLanguage("por")
	defer client.Close()

	client.SetImage(imagePath)
	text, err := client.Text()
	if err != nil {
		log.Printf("Erro ao realizar OCR: %v", err)
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

func CreateOutputDir() (string, error) {
	// Pasta onde a captura de tela será salva
	outputDir := "./output_screenshots"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err = os.Mkdir(outputDir, os.ModePerm)
		if err != nil {
			log.Fatalf("Erro ao criar diretório: %v", err)
			return "", err
		}
	}
	return outputDir, nil
}
