# Monitoramento Basico

## Endpoints

- `GET /api/health`
  - publico
  - serve para health check e liveness
- `GET /api/ready`
  - publico
  - indica se o runtime principal do painel subiu pronto
- `GET /api/ops/summary`
  - protegido
  - entrega contagem de usuarios, salas, mensagens, eventos, enquetes, sessoes, uploads, uptime e tamanho do banco

## Autorizacao do endpoint ops

- loopback local acessa sem token
- acesso remoto exige `Authorization: Bearer <UNIVERSALD_OPS_TOKEN>`

## Script local

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\check-ops.ps1 -BaseUrl http://127.0.0.1:7788
```

## Script remoto

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\check-ops.ps1 -BaseUrl https://teu-endereco -Token teu-token
```
