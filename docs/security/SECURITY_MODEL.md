# Security Model

## 1. Superficie publica

- `GET /api/health` eh publico para health check de hospedagem.
- Frontend estatico eh publico.
- Todo o resto do painel exige sessao valida.

## 2. Autenticacao e sessao

- Login por usuario ou email.
- Cookie `HttpOnly`, `SameSite=Lax` e `Secure` quando o acesso chega por HTTPS.
- Sessao invalida eh limpa no servidor quando falha autenticacao.
- Burst de senha errada entra em cooldown para reduzir abuso.

## 3. Autorizacao

- Cargos e permissoes resolvidos no backend.
- Salas ocultas, VIP, admin e salas com senha validam acesso no servidor.
- DMs, bloqueio, mute, logs e terminal respeitam permissao real de usuario.

## 4. Upload e midia

- Validacao por extensao, tipo e limite por cargo.
- Arquivos perigosos nao sao executados na mesma origem quando servidos por `/uploads/`.
- URLs de avatar e anexos passam por validacao e sanitizacao.

## 5. Headers e browser hardening

- CSP
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: SAMEORIGIN`
- `Referrer-Policy`
- `Permissions-Policy`
- `COOP` e `CORP`
- `HSTS` quando a conexao eh segura

## 6. Auditoria e recuperacao

- Logs administrativos persistidos.
- Backup operacional do banco e uploads por script.
- Smoke script para regressao rapida antes e depois de deploy.

## 7. Postura atual

O runtime padrao do projeto foi reduzido para o modo painel. Rotas legadas de app desktop nao ficam mais expostas no servidor web principal do `Painel Dief`.
