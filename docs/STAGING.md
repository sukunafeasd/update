# Painel Dief em Staging

## Objetivo

Manter uma copia publica separada do `Painel Dief` para validar ajustes sem mexer na URL principal de producao.

## Regras

- usar outra base de dados
- usar outro diretorio de uploads
- usar outro subdominio
- usar outro `UNIVERSALD_OPS_TOKEN`
- nunca apontar staging para o disco de producao

## Variaveis sugeridas

- `UNIVERSALD_APP_ENV=staging`
- `UNIVERSALD_PUBLIC_ORIGIN=https://staging.paineldief.example.com`
- `UNIVERSALD_DATA_DIR=/var/staging-data`
- `UNIVERSALD_DB=/var/staging-data/universald.db`
- `UNIVERSALD_UPLOADS=/var/staging-data/panel_uploads`
- `UNIVERSALD_OPS_TOKEN=troca-esse-token`

Blueprint inicial:

- [render.staging.yaml](/C:/Users/cafe/Desktop/site.dief/render.staging.yaml)

## Validacao minima antes de promover

1. `go test ./...`
2. smoke sem escrita
3. smoke com mutacao controlada
4. check do endpoint `/api/ops/summary`
5. validacao visual em desktop e celular
