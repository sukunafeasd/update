# Painel Dief

Plataforma privada online para grupo fechado, com chat em tempo real, DMs, perfis, uploads, areas protegidas, IA `Nego Dramias`, painel admin e biblioteca de midia persistente.

URL oficial atual:

- `https://update-wrl6.onrender.com`

## O que esta ativo

- Login por usuario ou email com sessao por cookie.
- 10 areas principais na lateral com foco em uso estilo app.
- Chat em tempo real com resposta, reacao, fixado, favorito, busca e digitando.
- DMs, salas privadas, salas com senha, cargos e controle de acesso.
- Upload persistente de imagem, video, audio e arquivo com preview e download.
- Perfil customizavel com avatar, bio, tema, status e status personalizado.
- Eventos com RSVP, enquetes, dashboard vivo, Apps Lab e IA integrada.
- Logs administrativos, anti-flood, limitador de login e headers de seguranca.

## Estrutura principal

- `cmd/universald`: servidor web do Painel Dief.
- `internal/api`: API HTTP e handlers do painel.
- `internal/panel`: regras de negocio do sistema social.
- `internal/db`: persistencia SQLite e migracoes.
- `web`: frontend do painel.
- `scripts`: smoke, backup e operacao local/publica.
- `scripts`: smoke, backup local, backup remoto da URL fixa e operacao.
- `docs`: arquitetura, producao e modelo de seguranca atual.

## Rodar localmente

```powershell
cd C:\Users\cafe\Desktop\site.dief
go mod tidy
go run .\cmd\universald
```

Endereco padrao:

- `http://127.0.0.1:7788`

## Variaveis uteis

- `PORT`
- `UNIVERSALD_BIND`
- `UNIVERSALD_DATA_DIR`
- `UNIVERSALD_DB`
- `UNIVERSALD_UPLOADS`
- `UNIVERSALD_WEB`
- `UNIVERSALD_APP_ENV`
- `UNIVERSALD_PUBLIC_ORIGIN`
- `UNIVERSALD_OPS_TOKEN`
- `UNIVERSALD_BACKUP_RETENTION_DAYS`
- `UNIVERSALD_MAINTENANCE_INTERVAL_SEC`
- `UNIVERSALD_OPEN=false`
- `UNIVERSALD_SAFE_MODE=true`

## Teste e operacao

Smoke sem escrita:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\smoke-panel.ps1 -BaseUrl http://127.0.0.1:7788
```

Smoke com mutacao controlada:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\smoke-panel.ps1 -BaseUrl http://127.0.0.1:7788 -MutatingChecks
```

Backup de banco e uploads:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\backup-panel.ps1
```

Backup remoto da URL oficial:

```powershell
$env:UNIVERSALD_OPS_TOKEN="teu-token"
powershell -ExecutionPolicy Bypass -File .\scripts\backup-remote-panel.ps1
```

Restore de backup:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\restore-panel.ps1 -ArchivePath .\backups\painel-dief-backup-YYYYMMDD-HHMMSS.zip -Force
```

Espelho local da producao gratis:

```powershell
$env:UNIVERSALD_OPS_TOKEN="teu-token"
powershell -ExecutionPolicy Bypass -File .\scripts\sync-production-mirror.ps1
```

Smoke da URL fixa:

```powershell
$env:UNIVERSALD_OPS_TOKEN="teu-token"
powershell -ExecutionPolicy Bypass -File .\scripts\smoke-production.ps1
```

Resumo operacional:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\check-ops.ps1 -BaseUrl http://127.0.0.1:7788
```

Preflight para URL fixa:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\deploy-readiness.ps1 -Provider render -PublicOrigin https://paineldief.example.com -Domain paineldief.example.com
```

Publicacao e URL fixa:

- [docs/PRODUCTION.md](/C:/Users/cafe/Desktop/site.dief/docs/PRODUCTION.md)
- [docs/STAGING.md](/C:/Users/cafe/Desktop/site.dief/docs/STAGING.md)
- [docs/MONITORING.md](/C:/Users/cafe/Desktop/site.dief/docs/MONITORING.md)
- [docs/FREE_PLAN.md](/C:/Users/cafe/Desktop/site.dief/docs/FREE_PLAN.md)

## Estado atual

O projeto esta consolidado como runtime do `Painel Dief`. Rotas antigas do desktop legado nao fazem mais parte do modo padrao do servidor web do site.

No plano gratis do Render, a URL fixa esta funcionando, mas o disco do runtime nao deve ser tratado como persistencia forte. Por isso a operacao correta e manter backup remoto e espelho local.
