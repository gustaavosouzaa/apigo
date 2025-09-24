# API de Geocodificação em Go

Esta API em Go expõe um endpoint REST para converter endereços em coordenadas de latitude e longitude utilizando o Google Maps Geocoding API. O projeto prioriza desempenho por meio de cache em memória e configurações de servidor HTTP otimizadas.

## Requisitos

- Go 1.21+
- Conta no Google Cloud com o Geocoding API habilitado

## Configuração

1. Instale as dependências:

   ```bash
   go mod tidy
   ```

2. Copie o arquivo `.env.example` para `.env` e configure a chave de API do Google Maps:

   ```bash
   cp .env.example .env
   # Edite o arquivo .env e informe sua chave
   ```

   Variáveis disponíveis:

   - `GOOGLE_MAPS_API_KEY` (obrigatória): chave de acesso ao Google Maps Geocoding API.
   - `PORT` (opcional, padrão `8080`): porta HTTP que o servidor irá escutar.

## Execução

Inicie o servidor:

```bash
go run ./...
```

A API ficará acessível em `http://localhost:8080` (ou na porta definida em `PORT`).

### Endpoints

- `GET /geocode?address=<endereco>`: retorna um JSON contendo o endereço formatado, latitude, longitude e a origem da informação (`google` ou `cache`).
- `GET /healthz`: endpoint de verificação simples que retorna o status `ok`.

### Exemplo de resposta

```json
{
  "address": "Praça da Sé - Sé, São Paulo - SP, 01001-000, Brasil",
  "latitude": -23.5505191,
  "longitude": -46.6333094,
  "source": "google"
}
```

## Observações de desempenho

- Resultados de geocodificação são armazenados em cache em memória por 30 minutos, reduzindo chamadas repetidas ao Google Maps e aumentando a capacidade de atendimento simultâneo.
- O servidor HTTP utiliza timeouts agressivos e cliente HTTP com timeout para evitar que requisições lentas degradem o serviço.

## Testes

Execute os testes (caso existam) com:

```bash
go test ./...
```
