# Painel Dief em Producao

## Objetivo

Publicar o `Painel Dief` com URL fixa, HTTPS, banco persistente, uploads persistentes e fluxo de atualizacao continua sem trocar o endereco do sistema.

## O que a base atual suporta

- bind automatico em `0.0.0.0:$PORT`
- runtime headless em host
- health check publico em `/api/health`
- banco e uploads fora do codigo via:
  - `UNIVERSALD_DATA_DIR`
  - `UNIVERSALD_DB`
  - `UNIVERSALD_UPLOADS`
- ambiente explicito por `UNIVERSALD_APP_ENV`
- origem publica explicita por `UNIVERSALD_PUBLIC_ORIGIN`
- endpoint de readiness em `/api/ready`
- endpoint operacional protegido em `/api/ops/summary`
- deploy por container com [Dockerfile](/C:/Users/cafe/Desktop/site.dief/Dockerfile)
- blueprint inicial para Render em [render.yaml](/C:/Users/cafe/Desktop/site.dief/render.yaml)
- smoke de regressao em [scripts/smoke-panel.ps1](/C:/Users/cafe/Desktop/site.dief/scripts/smoke-panel.ps1)
- backup simples em [scripts/backup-panel.ps1](/C:/Users/cafe/Desktop/site.dief/scripts/backup-panel.ps1)
- restore de backup em [scripts/restore-panel.ps1](/C:/Users/cafe/Desktop/site.dief/scripts/restore-panel.ps1)
- consulta operacional em [scripts/check-ops.ps1](/C:/Users/cafe/Desktop/site.dief/scripts/check-ops.ps1)
- preflight de URL fixa em [scripts/deploy-readiness.ps1](/C:/Users/cafe/Desktop/site.dief/scripts/deploy-readiness.ps1)
- export remoto protegido em `/api/ops/export`
- backup remoto em [scripts/backup-remote-panel.ps1](/C:/Users/cafe/Desktop/site.dief/scripts/backup-remote-panel.ps1)
- espelho local da producao em [scripts/sync-production-mirror.ps1](/C:/Users/cafe/Desktop/site.dief/scripts/sync-production-mirror.ps1)
- smoke oficial em [scripts/smoke-production.ps1](/C:/Users/cafe/Desktop/site.dief/scripts/smoke-production.ps1)

## Variaveis principais

- `PORT`
- `UNIVERSALD_APP_ENV=production`
- `UNIVERSALD_PUBLIC_ORIGIN=https://paineldief.example.com`
- `UNIVERSALD_DATA_DIR=/var/data`
- `UNIVERSALD_DB=/var/data/universald.db`
- `UNIVERSALD_UPLOADS=/var/data/panel_uploads`
- `UNIVERSALD_WEB=/app/web`
- `UNIVERSALD_OPEN=false`
- `UNIVERSALD_SAFE_MODE=true`
- `UNIVERSALD_OPS_TOKEN=troca-esse-token`
- `UNIVERSALD_BACKUP_RETENTION_DAYS=14`
- `UNIVERSALD_MAINTENANCE_INTERVAL_SEC=60`

Na blueprint atual da Render, `UNIVERSALD_OPS_TOKEN` ja pode ser gerado automaticamente no primeiro deploy.

## Fluxo recomendado

1. Subir o repositorio em um host Git.
2. Publicar o web service com o `Dockerfile`.
3. Montar disco persistente em `/var/data`.
4. Configurar o health check em `/api/health` e readiness em `/api/ready`.
5. Rodar o smoke script contra a URL publicada.
6. Configurar `UNIVERSALD_PUBLIC_ORIGIN` com a URL final.
7. Definir `UNIVERSALD_OPS_TOKEN` e guardar fora do repo.
8. Habilitar dominio proprio quando quiser sair do subdominio do provedor.

## O que ainda falta fora do codigo para URL fixa

- um repositorio Git hospedado
- um provedor com deploy web e disco persistente
- um dominio ou subdominio com acesso DNS
- um `UNIVERSALD_OPS_TOKEN` real guardado fora do repo

## Preflight rapido

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\deploy-readiness.ps1 -Provider render -PublicOrigin https://paineldief.example.com -Domain paineldief.example.com
```

## Smoke

Sem escrita:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\smoke-panel.ps1 -BaseUrl https://teu-endereco
```

Com mutacao controlada:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\smoke-panel.ps1 -BaseUrl https://teu-endereco -MutatingChecks
```

## Backup

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\backup-panel.ps1
```

O backup agora inclui `backup-manifest.json` e pode podar zips antigos:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\backup-panel.ps1 -RetentionDays 14
```

## Restore

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\restore-panel.ps1 -ArchivePath .\backups\painel-dief-backup-YYYYMMDD-HHMMSS.zip -Force
```

## Operacao e monitoramento

Resumo protegido:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\check-ops.ps1 -BaseUrl https://teu-endereco -Token teu-token
```

Export de backup remoto:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\backup-remote-panel.ps1 -BaseUrl https://teu-endereco -OpsToken teu-token
```

## Plano gratis

Se o host estiver no plano gratis sem disco persistente:

- trata a URL fixa como producao oficial
- nao trata o runtime como armazenamento confiavel
- mantem backup remoto fora do provedor
- mantem espelho local restauravel
- roda smoke recorrente antes e depois de atualizar

## Limite deste ambiente

O projeto esta pronto para producao fixa, mas a publicacao permanente ainda depende das credenciais da plataforma de hospedagem e, se houver, do dominio final.
