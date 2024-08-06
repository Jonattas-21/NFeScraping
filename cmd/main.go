package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"nfe-scraping/domain"
	"nfe-scraping/infrastructure"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
)

// Function to extract and save images of invoices from a PDF file
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
			text, err := infrastructure.PerformOCR(imagePath)

			if err != nil {
				log.Printf("Error in OCR execution %s: %v", imagePath, err)
				continue
			}

			if domain.IsNotaFiscal(text) {
				nfe := domain.NotaFiscal{FullText: text}
				nfe.ExtractCNPJ()
				nfe.ExtractCodigoVerificacao()
				nfe.ExtractMunicipio()
				nfe.ExtractNumeroNota()
				nfe.CorectExtractNotaFiscal()
				if nfe.CorectExtract {
					nfe.ScrapingNotaFiscalSP()
				} else {
					nfe.Status = "NÃO PROCESSADA CORRETAMENTE"
				}
				nfe.DocumentPage, _ = infrastructure.ExtractPageFromFileImageName(imagePath)
				FoundNotaFiscal = append(FoundNotaFiscal, nfe)

			} else {
				// Delete image file if it is not an invoice
				os.Remove(imagePath)
			}
		}
	}

	return nil
}

var FoundNotaFiscal []domain.NotaFiscal

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	pdfPath := os.Getenv("PDFINPUTPATH")

	defer infrastructure.CleanUp()

	// Extrair e salvar imagens de notas fiscais
	err = extractAndSaveNotaFiscalImages(pdfPath)

	colunsForExcel := domain.PrepareNfeColumnsForExcel()
	var excelNfeVCalues [][]string
	for _, nfe := range FoundNotaFiscal {
		fmt.Println("========================= NF ===================================")
		fmt.Printf("CNPJ: %s\n", nfe.CNPJ)
		fmt.Printf("Código de Verificação: %s\n", nfe.CodigoVerificacao)
		fmt.Printf("Município: %s\n", nfe.Municipio)
		fmt.Printf("Número da Nota: %s\n", nfe.NumeroNota)
		fmt.Printf("Extração correta: %t\n", nfe.CorectExtract)
		fmt.Printf("Status: %s\n", nfe.Status)

		//Build excel output
		infrastructure.CreateExcelOutputDir()
		excelNfeVCalues = append(excelNfeVCalues, nfe.PrepareForExcel())
	}

	infrastructure.CreateExcelOutput(colunsForExcel, excelNfeVCalues)

	if err != nil {
		log.Fatalf("error extracting images from PDF: %v", err)
	}
}
