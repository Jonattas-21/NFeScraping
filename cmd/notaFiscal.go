package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
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
		nfe.CodigoVerificacao = match[1]
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

func (nfe *NotaFiscal) IsValidNotaFiscal() (bool, error) {
	text, err := PerformOCR(nfe.ScreenshotPath)
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

	if err != nil {
		log.Printf("Erro ao executar OCR na imagem %s: %v", nfe.ScreenshotPath, err)
		return false, err
	}

	nfe.Status = "VÁLIDA"
	return true, nil
}

func (nfe *NotaFiscal) ScrapingNotaFiscalSP() error {
	fmt.Println("Começou o scraping.")

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// URL da página de verificação
	url := "https://nfe.prefeitura.sp.gov.br/publico/verificacao.aspx"

	fmt.Println("Iniciando scraping para a nota fiscal:", nfe.NumeroNota)

	// Variável para armazenar o resultado extraído
	//var result string = "teste"
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
		log.Fatalf("Erro ao executar chromedp: %v", err)
		return err
	}

	log.Println("Resultado extraído com sucesso, agora vamos baixar a imagem.")
	// Pasta onde a captura de tela será salva
	outputDir, err := CreateOutputDir()
	if err != nil {
		return err
	}

	// Caminho completo para salvar a captura de tela
	filePath := fmt.Sprintf("%s/%s_print.png", outputDir, nfe.NumeroNota)
	nfe.ScreenshotPath = filePath

	// Salvar captura de tela
	err = ioutil.WriteFile(filePath, screenshot, 0644)
	if err != nil {
		log.Fatalf("Erro ao salvar captura de tela: %v", err)
		return err
	}

	nfe.IsValidNotaFiscal()

	return nil
}
