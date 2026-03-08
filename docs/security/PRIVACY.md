# Privacy

## Dados que o painel guarda

- usuarios, cargos e sessoes
- mensagens e historico de chat
- eventos, enquetes, favoritos e pins
- uploads e metadados de arquivo
- logs administrativos do proprio painel

## O que nao entra por padrao

- nenhuma sincronizacao com nuvem por conta propria
- nenhum envio de telemetria externa por padrao
- nenhum compartilhamento com terceiro fora do host escolhido

## Retencao

- mensagens e arquivos ficam persistidos ate limpeza explicita
- logs seguem a politica local do banco
- backups dependem da rotina operacional definida pelo admin

## Boas praticas operacionais

- usar disco persistente
- fazer backup periodico de `universald.db` e `panel_uploads`
- restringir acesso ao ambiente de hospedagem
- proteger o dominio e a conta do provedor
