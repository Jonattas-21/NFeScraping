package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/otiai10/gosseract/v2"
)

// Função para verificar se o texto contém padrões de uma nota fiscal
func isNotaFiscal(text string) bool {

	cnpjPattern := regexp.MustCompile(`\b\d{2}\.\d{3}\.\d{3}/\d{4}-\d{2}\b`)
	dataPattern := regexp.MustCompile(`\b\d{2}/\d{2}/\d{4}\b`)
	valorPattern := regexp.MustCompile(`R\$ ?\d+,\d{2}`)
	palavrasChave := []string{"NOTA FISCAL", "NFS-e"}

	if cnpjPattern.MatchString(text) && dataPattern.MatchString(text) && valorPattern.MatchString(text) {
		return true
	}

	count := len(palavrasChave)
	for _, palavra := range palavrasChave {
		if strings.Contains(strings.ToUpper(text), strings.ToUpper(palavra)) {
			count--
			log.Println("Palavra chave encontrada:", palavra)
		}
	}

	return count == 0
}

// Função para realizar OCR em uma imagem
func performOCR(imagePath string) (string, error) {
	client := gosseract.NewClient()
	defer client.Close()

	client.SetImage(imagePath)
	text, err := client.Text()
	if err != nil {
		log.Printf("Erro ao realizar OCR: %v", err)
		return "", err
	}

	return text, nil
}

// Função para extrair o CNPJ
func extractCNPJ(text string) string {
	cnpjPattern := regexp.MustCompile(`CPF/CNPJ:\s*([\d\.\-\/]+)`)
	match := cnpjPattern.FindStringSubmatch(text)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

// Função para extrair o código de verificação
func extractCodigoVerificacao(text string) string {
	// Padrão específico para capturar códigos como ILC2-I9PB
	codigoPattern := regexp.MustCompile(`(?:Código de Verificação|Codigo de Verificacao|Códig de Verificagao|Cédigo de Verificagao|Cédig de Verificagao)\s*([A-Z0-9\-]+)`)
	matches := codigoPattern.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		if len(match) > 1 {
			return match[1]
		}
	}

	// Se não encontrado na linha atual, procure na linha seguinte
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.Contains(line, "Código de Verificação") && i+1 < len(lines) {
			nextLinePattern := regexp.MustCompile(`[A-Z0-9\-]+`)
			nextLineMatch := nextLinePattern.FindString(lines[i+1])
			if nextLineMatch != "" {
				return nextLineMatch
			}
		}
	}

	return ""
}

// Função para extrair o município
func extractMunicipio(text string) string {
	municipioPattern := regexp.MustCompile(`Municipio:\s*([A-Za-z\s]+)\s+UF`)
	match := municipioPattern.FindStringSubmatch(text)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

// Função para extrair o número da nota
func extractNumeroNota(text string) string {
	numeroPattern := regexp.MustCompile(`\b\d{8}\b`)
	match := numeroPattern.FindString(text)
	return match
}

// Função para extrair e salvar imagens de notas fiscais
func extractAndSaveNotaFiscalImages(pdfPath string) ([]string, error) {
	outputDir := "./output_images"
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return nil, err
	}

	cmd := exec.Command("pdfimages", "-j", pdfPath, outputDir+"/image")
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(outputDir)
	if err != nil {
		return nil, err
	}

	var notaFiscalImages []string
	for _, file := range files {
		if !file.IsDir() {
			imagePath := outputDir + "/" + file.Name()
			text, err := performOCR(imagePath)
			if err != nil {
				log.Printf("Erro ao executar OCR na imagem %s: %v", imagePath, err)
				continue
			}

			if isNotaFiscal(text) {
				notaFiscalImages = append(notaFiscalImages, imagePath)

				// Extrair e imprimir informações da nota fiscal
				cnpj := extractCNPJ(text)
				codigoVerificacao := extractCodigoVerificacao(text)
				municipio := extractMunicipio(text)
				numeroNota := extractNumeroNota(text)

				log.Println("==============================================================")
				fmt.Printf("Imagem %s contém uma Nota Fiscal.\n", imagePath)
				fmt.Printf("CNPJ: %s\n", cnpj)
				fmt.Printf("Código de Verificação: %s\n", codigoVerificacao)
				fmt.Printf("Município: %s\n", municipio)
				fmt.Printf("Número da Nota: %s\n", numeroNota)
				log.Println("==============================================================")
				fmt.Printf("Texto extraído:\n%s\n", text)

			} else {
				// Se não for uma nota fiscal, pode optar por excluir a imagem
				os.Remove(imagePath)
			}
		}
	}

	return notaFiscalImages, nil
}

func main() {
	pdfPath := "./input/teste-1.pdf"

	// Extrair e salvar imagens de notas fiscais
	images, err := extractAndSaveNotaFiscalImages(pdfPath)
	if err != nil {
		log.Fatalf("Erro ao extrair imagens do PDF: %v", err)
	}

	for _, imagePath := range images {
		fmt.Printf("Imagem %s contém uma Nota Fiscal.\n", imagePath)
	}
}
