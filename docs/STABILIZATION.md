# Stabilization Pass

Rodada focada em consolidar o `Painel Dief` e o `universalD` sem abrir feature grande nova.

## Foco

- persistencia de sessao com fallback local
- persistencia de rascunho por sala
- restauracao da sala ativa por usuario
- paridade visual e operacional entre site e app
- smoke rapido do desktop nativo
- QA pass unico para build + smoke

## Checklist rapido

```powershell
cd C:\Users\cafe\Desktop\site.dief
powershell -ExecutionPolicy Bypass -File .\scripts\qa-pass.ps1
```

## Desktop

Smoke dedicado do exe:

```powershell
powershell -ExecutionPolicy Bypass -File C:\Users\cafe\Desktop\universalD\scripts\smoke-desktop.ps1
```

## Objetivo

Manter o produto mais estavel, previsivel e polido antes de qualquer expansao grande de escopo.
