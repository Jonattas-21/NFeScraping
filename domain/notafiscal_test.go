package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ExtractCNPJ_ok(t *testing.T) {
	nfe := NotaFiscal{FullText: "PRESTADOR DE SERVIÇOS\nCPF/CNPJ: 12.345.678/0001-90"}
	nfe.ExtractCNPJ()
	if nfe.CNPJ != "12.345.678/0001-90" {
		t.Errorf("CNPJ was not extracted correctly")
	}
}

func Test_ExtractCNPJ_error(t *testing.T) {
	assert := assert.New(t)
	nfe := NotaFiscal{FullText: "TOMADOR DE SERVIÇOS\nCPF/CNPJ: 12.345.678/0001-90"}
	nfe.ExtractCNPJ()

	assert.Equal("", nfe.CNPJ, "CNPJ was not extracted correctly")
}

func Test_ExtractCodigoVerificacao_ok(t *testing.T) {
	nfe := NotaFiscal{FullText: "Código de Verificação \n: 1234-5AW8"}
	nfe.ExtractCodigoVerificacao()
	if nfe.CodigoVerificacao != "1234-5AW8" {
		t.Errorf("verify code was not extracted correctly")
	}
}

func Test_ExtractCodigoVerificacao_error(t *testing.T) {
	assert := assert.New(t)
	nfe := NotaFiscal{FullText: "Código de Verificação \n: 1234-5AW"}
	nfe.ExtractCodigoVerificacao()

	assert.Equal("", nfe.CodigoVerificacao, "verify code was not extracted correctly")
}

func Test_ExtractMunicipio_ok(t *testing.T) {
	nfe := NotaFiscal{FullText: "Município: São Paulo UF"}
	nfe.ExtractMunicipio()
	if nfe.Municipio != "São Paulo" {
		t.Errorf("Município was not extracted correctly")
	}
}

func Test_ExtractMunicipio_error(t *testing.T) {
	assert := assert.New(t)
	nfe := NotaFiscal{FullText: "Município: São Paulo"}
	nfe.ExtractMunicipio()

	assert.Equal("", nfe.Municipio, "Município was not extracted correctly")
}

func Test_ExtractNumeroNota_ok(t *testing.T) {
	nfe := NotaFiscal{FullText: "Número da Nota: 12345678"}
	nfe.ExtractNumeroNota()
	if nfe.NumeroNota != "12345678" {
		t.Errorf("Número da Nota was not extracted correctly")
	}
}

func Test_ExtractNumeroNota_error(t *testing.T) {
	assert := assert.New(t)
	nfe := NotaFiscal{FullText: "Número da Nota: 123456x8"}
	nfe.ExtractNumeroNota()

	assert.Equal("", nfe.NumeroNota, "Número da Nota was not extracted correctly")
}

func Test_CorectExtractNotaFiscal_ok(t *testing.T) {
	nfe := NotaFiscal{NumeroNota: "12345678", CNPJ: "12.345.678/0001-90", CodigoVerificacao: "1234-5AW8"}
	nfe.CorectExtractNotaFiscal()
	if !nfe.CorectExtract {
		t.Errorf("Nota Fiscal was not extracted correctly")
	}
}

func Test_CorectExtractNotaFiscal_error(t *testing.T) {
	assert := assert.New(t)

	// Test 1 - CodigoVerificacao is empty
	nfe := NotaFiscal{NumeroNota: "12345678", CNPJ: "12.345.678/0001-90", CodigoVerificacao: ""}
	nfe.CorectExtractNotaFiscal()
	assert.False(nfe.CorectExtract, "Nota Fiscal was not extracted correctly")

	// Test 2 - CNPJ is empty
	nfe = NotaFiscal{NumeroNota: "12345678", CNPJ: "", CodigoVerificacao: "1234-5AW8"}
	nfe.CorectExtractNotaFiscal()
	assert.False(nfe.CorectExtract, "Nota Fiscal was not extracted correctly")

	// Test 3 - NumeroNota is empty
	nfe = NotaFiscal{NumeroNota: "", CNPJ: "12.345.678/0001-90", CodigoVerificacao: "1234-5AW8"}
	nfe.CorectExtractNotaFiscal()
	assert.False(nfe.CorectExtract, "Nota Fiscal was not extracted correctly")
}

func Test_IsValidQueryNotaFiscall_ok(t *testing.T) {
	nfe := NotaFiscal{ScreenshotPath: "test.png"}
	result, err := nfe.IsValidQueryNotaFiscal()
	if !result {
		t.Errorf("Nota Fiscal was not validated correctly")
	}
	if err != nil {
		t.Errorf("Error was not expected")
	}
}
