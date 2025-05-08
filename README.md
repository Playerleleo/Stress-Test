# Stress Test CLI

Uma ferramenta CLI em Go para realizar testes de carga em serviços web.

## Requisitos

- Go 1.21 ou superior
- Docker (opcional)

## Como Usar

### Executando Localmente

```bash
go run main.go --url=<URL> --requests=<N> --concurrency=<N>
```

### Executando com Docker

1. Construa a imagem:
```bash
docker build -t stress-test .
```

2. Execute o container:
```bash
docker run stress-test --url=<URL> --requests=<N> --concurrency=<N>
```

## Parâmetros

- `--url`: URL do serviço a ser testado (obrigatório)
- `--requests`: Número total de requests (obrigatório)
- `--concurrency`: Número de chamadas simultâneas (obrigatório)

## Exemplo

```bash
docker run stress-test --url=http://google.com --requests=1000 --concurrency=10
```

## Relatório

O sistema gera um relatório contendo:
- Tempo total de execução
- Total de requests realizados
- Quantidade de requests com sucesso (status 200)
- Distribuição de códigos de status HTTP