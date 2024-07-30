package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
