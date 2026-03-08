# Painel Dief no Plano Gratis

## Premissa

Sem gastar com disco persistente, a URL fixa pode ficar publica e funcional, mas o armazenamento do host nao deve ser tratado como confiavel para longo prazo.

URL oficial atual:

- `https://update-wrl6.onrender.com`

## Operacao minima correta

1. manter `UNIVERSALD_OPS_TOKEN` salvo fora do projeto
2. rodar backup remoto com [backup-remote-panel.ps1](/C:/Users/cafe/Desktop/site.dief/scripts/backup-remote-panel.ps1)
3. manter um espelho local via [sync-production-mirror.ps1](/C:/Users/cafe/Desktop/site.dief/scripts/sync-production-mirror.ps1)
4. rodar [smoke-production.ps1](/C:/Users/cafe/Desktop/site.dief/scripts/smoke-production.ps1) antes e depois de atualizacao
5. evitar abrir escopo grande sem necessidade

## Fluxo sugerido

Backup remoto:

```powershell
$env:UNIVERSALD_OPS_TOKEN="teu-token"
powershell -ExecutionPolicy Bypass -File .\scripts\backup-remote-panel.ps1
```

Espelho local:

```powershell
$env:UNIVERSALD_OPS_TOKEN="teu-token"
powershell -ExecutionPolicy Bypass -File .\scripts\sync-production-mirror.ps1
```

Smoke oficial:

```powershell
$env:UNIVERSALD_OPS_TOKEN="teu-token"
powershell -ExecutionPolicy Bypass -File .\scripts\smoke-production.ps1
```

## Limite real

Sem disco persistente no provedor, a recuperacao total da producao depende desses backups externos.
