# Arquitetura Painel Dief

## Visao geral

O `Painel Dief` roda como uma aplicacao web com backend Go, frontend estatico servido pelo proprio processo e persistencia local em SQLite.

Fluxo principal:

1. O navegador carrega `web/index.html`, `web/styles.css` e `web/app.js`.
2. O frontend autentica em `/api/panel/login` e recebe sessao por cookie.
3. O cliente trabalha em cima de rotas `/api/panel/*` e stream SSE em `/api/panel/stream`.
4. Dados persistem em SQLite e arquivos ficam em `panel_uploads`.

## Camadas

### `cmd/universald`

- Bootstrap do servidor HTTP.
- Carrega configuracao e diretorios persistentes.
- Inicializa o servico do painel.

### `internal/api`

- Exposicao das rotas HTTP do Painel Dief.
- Auth por sessao de painel.
- Headers de seguranca, health check, serving estatico e uploads seguros.

### `internal/panel`

- Regras de negocio do produto.
- Login, sessao, cargos, presenca, salas, mensagens, DMs, IA, enquetes e eventos.
- Validacao de upload, busca, favoritos, pins, bloqueio e mute.

### `internal/db`

- Banco SQLite.
- Migracoes e queries do painel.
- Persistencia de historico, usuarios, salas, mensagens, enquetes, eventos e logs.

### `web`

- Login e shell principal do Painel Dief.
- Sidebar com 10 areas, chat principal e painel de apoio.
- Composer, biblioteca de midia, perfis, dashboard e Apps Lab.

## Estrutura de dados

- `universald.db`: banco principal do painel.
- `panel_uploads/`: midia e anexos persistidos.
- `.runtime/`: pids e logs operacionais locais.

## Tempo real

- SSE para atualizacao de sala, presenca, notificacao e novidades.
- Polling leve apenas onde faz sentido para reforco de UX.

## Seguranca

- Senhas com hash no backend.
- Cookie `HttpOnly`, `SameSite=Lax` e `Secure` quando a requisicao eh HTTPS.
- Sanitizacao de perfil e URLs de avatar.
- Upload servido com protecao contra conteudo executavel na mesma origem.
- Health check publico, resto protegido por sessao.

## Direcao atual

O projeto esta em fase de consolidacao:

- menos escopo novo
- mais estabilidade
- mais cobertura de teste
- mais limpeza de legado
- deploy fixo e atualizacao continua sem mudar a URL
