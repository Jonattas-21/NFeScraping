package domain

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"nfe-scraping/infrastructure"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type NotaFiscal struct {
	CNPJ              string
	CodigoVerificacao string
	Municipio         string
	NumeroNota        string
	FullText          string
	ScreenshotPath    string
	Status            string
	CorectExtract     bool
	RequestResult     string
	DocumentPage      int
}

// Função para extrair o CNPJ
func (nfe *NotaFiscal) ExtractCNPJ() {
	lines := strings.Split(nfe.FullText, "\n")
	for i, line := range lines {
		if strings.Contains(strings.ToUpper(line), "PRESTADOR DE SERVIÇOS") {
			if i+1 < len(lines) {
				// Atualizar o padrão de regex para capturar apenas o CNPJ
				cnpjPattern := regexp.MustCompile(`(?:CPF\/CNPJ:?\s*|\bCNPJ\b\s*)([\d\.\-\/]+)`)
				match := cnpjPattern.FindStringSubmatch(lines[i+1])
				if len(match) > 1 {
					//return match[1]
					nfe.CNPJ = match[1]
				}
			}
		}
	}
}

// Função para extrair o código de verificação
func (nfe *NotaFiscal) ExtractCodigoVerificacao() {
	codigoPattern := regexp.MustCompile(`(?:Código de Verificação|Codigo de Verificacao|Códig de Verificagao|Cédigo de Verificagao|Cédig de Verificagao)[^\n]*\n.*\b([A-Z0-9]{4}-[A-Z0-9]{4})\b`)
	match := codigoPattern.FindStringSubmatch(nfe.FullText)
	if len(match) > 1 {
		// Check if the code has at least 2 letters and 2 numbers
		if len(regexp.MustCompile(`[A-Z0-9]`).FindAllString(match[1], -1)) >= 2 {
			nfe.CodigoVerificacao = match[1]
		} else {
			log.Fatal("verify code has less than 2 letters and 2 numbers")
		}
	}
}

// Função para extrair o município
func (nfe *NotaFiscal) ExtractMunicipio() {
	// Atualizar a expressão regular para considerar letras acentuadas e espaços
	municipioPattern := regexp.MustCompile(`Município:\s*([\p{L}\p{Zs}]+)\s+UF`)
	match := municipioPattern.FindStringSubmatch(nfe.FullText)
	if len(match) > 1 {
		nfe.Municipio = match[1]
	}
}

// Função para extrair o número da nota
func (nfe *NotaFiscal) ExtractNumeroNota() {
	numeroPattern := regexp.MustCompile(`\b\d{8}\b`)
	match := numeroPattern.FindString(nfe.FullText)
	nfe.NumeroNota = match
}

func (nfe *NotaFiscal) CorectExtractNotaFiscal() {
	if nfe.NumeroNota != "" && nfe.CNPJ != "" && nfe.CodigoVerificacao != "" {
		nfe.CorectExtract = true
	} else {
		nfe.CorectExtract = false
	}
}

func (nfe *NotaFiscal) IsValidQueryNotaFiscal() (bool, error) {
	text, err := infrastructure.PerformOCR(nfe.ScreenshotPath)
	if err != nil {
		log.Printf("error executing ocr in image %s: %v", nfe.ScreenshotPath, err)
		return false, err
	}

	nfe.RequestResult = text

	if strings.Contains(text, "Número da NFS-e e Código de Verificação não conferem.") {
		nfe.Status = "CAMPOS INVALIDOS"
		return false, nil
	}

	stgCancelada := []string{"CANCELADA", "CANBELADA", "CANCELADA"}

	for _, stg := range stgCancelada {
		if strings.Contains(strings.ToUpper(text), stg) {
			nfe.Status = "CANCELADA"
			log.Println("Nota fiscal cancelada.")
			return false, nil
		}
	}

	nfe.Status = "VÁLIDA"
	return true, nil
}

func (nfe *NotaFiscal) ScrapingNotaFiscalSP() error {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// URL da página de verificação
	url := "https://nfe.prefeitura.sp.gov.br/publico/verificacao.aspx"

	fmt.Println("Scraping starts for NF:", nfe.NumeroNota)

	// Variável para armazenar o resultado extraído
	var screenshot []byte

	// nfe.CNPJ = "30.627.283/0001-67"
	// nfe.CodigoVerificacao = "K5PA-R7VG"
	// nfe.NumeroNota = "00000387"

	// Executar tarefas
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`#ctl00_body_tbCPFCNPJ`, chromedp.ByID),
		chromedp.SendKeys(`#ctl00_body_tbCPFCNPJ`, nfe.CNPJ, chromedp.ByID),
		chromedp.SendKeys(`#ctl00_body_tbNota`, nfe.NumeroNota, chromedp.ByID),
		chromedp.SendKeys(`#ctl00_body_tbVerificacao`, nfe.CodigoVerificacao, chromedp.ByID),
		chromedp.Click(`#ctl00_body_btVerificar`, chromedp.ByID),
		chromedp.Sleep(5*time.Second),
		chromedp.FullScreenshot(&screenshot, 100),
		//chromedp.OuterHTML(`#ctl00_cphBase_img`, &result, chromedp.ByID),
	)

	if err != nil {
		log.Fatalf("error executing chromedp: %v", err)
		return err
	}

	log.Println("Scraping was successful. Now lets save the screenshot.")
	// Pasta onde a captura de tela será salva
	outputDir, err := infrastructure.CreateOutputDir()
	if err != nil {
		return err
	}

	// Caminho completo para salvar a captura de tela
	filePath := fmt.Sprintf("%s/%s_print.png", outputDir, nfe.NumeroNota)
	nfe.ScreenshotPath = filePath

	// Salvar captura de tela
	err = ioutil.WriteFile(filePath, screenshot, 0644)
	if err != nil {
		log.Fatalf("error saving the screenshot: %v", err)
		return err
	}

	nfe.IsValidQueryNotaFiscal()

	return nil
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

func (nfe *NotaFiscal) PrepareForExcel() []string {

	var values []string

	values = append(values, nfe.CNPJ)
	values = append(values, nfe.CodigoVerificacao)
	values = append(values, nfe.NumeroNota)
	values = append(values, nfe.Municipio)
	values = append(values, fmt.Sprintf("%t", nfe.CorectExtract))
	values = append(values, nfe.Status)
	values = append(values, fmt.Sprintf("%d", nfe.DocumentPage))

	return values
}

func PrepareNfeColumnsForExcel() []string {
	return []string{
		"CNPJ tomador de serviço",
		"Código de verificação",
		"Número da nota",
		"Município",
		"Extraiu corretamente?",
		"Status da nota fiscal consultada",
		"Página no documento",
	}
}
