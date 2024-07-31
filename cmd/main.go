package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

// Função para extrair e salvar imagens de notas fiscais
func extractAndSaveNotaFiscalImages(pdfPath string) error {
	outputDir := "./output_images"
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return err
	}

	cmd := exec.Command("pdfimages", "-j", pdfPath, outputDir+"/image")
	if err := cmd.Run(); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(outputDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			imagePath := outputDir + "/" + file.Name()
			text, err := PerformOCR(imagePath)

			if err != nil {
				log.Printf("Erro ao executar OCR na imagem %s: %v", imagePath, err)
				continue
			}

			if IsNotaFiscal(text) {
				nfe := NotaFiscal{FullText: text}

				// Extrair e imprimir informações da nota fiscal
				nfe.ExtractCNPJ()
				nfe.ExtractCodigoVerificacao()
				nfe.ExtractMunicipio()
				nfe.ExtractNumeroNota()

				fmt.Printf("Imagem %s contém uma Nota Fiscal.\n", imagePath)
				fmt.Printf("CNPJ: %s\n", nfe.CNPJ)
				fmt.Printf("Código de Verificação: %s\n", nfe.CodigoVerificacao)
				fmt.Printf("Município: %s\n", nfe.Municipio)
				fmt.Printf("Número da Nota: %s\n", nfe.NumeroNota)
				fmt.Println("==============================================================")

				nfe.ScrapingNotaFiscalSP()
				//Nfes = append(Nfes, nfe)

				fmt.Printf("Imagem %s contém uma Nota Fiscal.\n", imagePath)

			} else {
				// Se não for uma nota fiscal, pode optar por excluir a imagem
				os.Remove(imagePath)
			}
		}
	}

	return nil
}

func main() {
	pdfPath := "./input/teste-1.pdf"

	// Extrair e salvar imagens de notas fiscais
	err := extractAndSaveNotaFiscalImages(pdfPath)
	if err != nil {
		log.Fatalf("Erro ao extrair imagens do PDF: %v", err)
	}
}
