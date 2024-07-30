# Etapa 1: Configuração do Ambiente e Instalação das Dependências
FROM ubuntu:22.04 AS build-env

# Instalar dependências do sistema e ferramentas necessárias
RUN apt-get update && apt-get install -y \
    tesseract-ocr \
    libleptonica-dev \
    libtesseract-dev \
    poppler-utils \
    wget \
    libpng-dev \
    libjpeg-dev \
    pkg-config \
    golang-go \
    && rm -rf /var/lib/apt/lists/*

# Verificar a instalação dos arquivos de cabeçalho do Tesseract e Leptonica
RUN find /usr/include -name baseapi.h
RUN find /usr/include -name allheaders.h
RUN find /usr/include -name leptonica

# Configurar o diretório de trabalho
WORKDIR /app

# Copiar arquivos de módulo do Go
COPY go.mod go.sum ./

# Baixar dependências do Go
RUN go mod download || { echo 'Erro ao baixar dependências do Go'; exit 1; }

# Copiar o código fonte
COPY . .

# Exibir informações sobre a estrutura do diretório de trabalho para diagnóstico
RUN ls -R /app

# Verificar se as dependências do Go estão corretamente baixadas
RUN go list -m all

# Compilar a aplicação
RUN go build -o app ./cmd/main.go || { echo 'Erro ao compilar a aplicação'; exit 1; }

# Etapa 2: Configuração do ambiente de execução
FROM ubuntu:22.04

# Instalar apenas o necessário para a execução
RUN apt-get update && apt-get install -y \
    tesseract-ocr \
    libpng-dev \
    libjpeg-dev \
    && rm -rf /var/lib/apt/lists/*

# Verificar a instalação dos arquivos de cabeçalho do Tesseract e Leptonica na imagem de execução
RUN find /usr/include -name baseapi.h
RUN find /usr/include -name allheaders.h
RUN find /usr/include -name leptonica

# Configurar variáveis de ambiente para o Tesseract
ENV TESSDATA_PREFIX=/usr/share/tesseract-ocr/4.00/tessdata/

# Criar diretório para a aplicação
WORKDIR /app

# Copiar o binário da aplicação do estágio de construção
COPY --from=build-env /app/app .

# Comando para executar a aplicação
CMD ["./app"]
