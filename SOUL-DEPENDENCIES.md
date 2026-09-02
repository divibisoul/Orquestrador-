# SOUL Dependencies — Frente 3

## N02 — CI / merge gate

O PR #13 do N02 já foi mesclado em 2026-09-02. Portanto, ele não pode mais ser bloqueado retroativamente. A Frente 3 cria o gate `.github/workflows/merge-gate.yml` no N02 atual para que novas alterações ao `main` sejam validadas por `npm install`, `npm run lint` e `npm run build` antes de aprovação. O workflow também mantém a sinalização `ci-pending` enquanto a validação da execução do PR não for bem-sucedida.

A proteção efetiva contra merge exige que o check do workflow seja marcado como obrigatório nas regras de proteção/regras do `main` no GitHub. O workflow, isoladamente, não consegue bloquear um merge de forma retroativa nem impedir um merge administrativo quando o check não foi configurado como obrigatório.

## N07 — Secrets obrigatórios

Configure no repositório `divibisoul/Orquestrador-`:

- `SOUL_MESH_TOKEN`: credencial canônica da Soul Mesh.
- `REPO_SECRETS_ADMIN_TOKEN`: credencial separada com permissão mínima para escrever Actions secrets nos seis destinos.
- `SOUL_MESH_N01_URL` até `SOUL_MESH_N06_URL`: endpoints reais dos seis núcleos para o validador E2E.
- `N07_APP_TOKEN`: token de aplicação do dashboard.

Nunca coloque qualquer valor de secret em código, YAML literal, issues, PRs, logs ou artefatos.

## N07 — Peers e dashboard

`api/health/dashboard.go` mantém autenticação `N07_APP_TOKEN`, preserva histórico e passou a materializar sempre N01–N06 no inventário. Sem `SOUL_MESH_PEERS`, cada peer fica explicitamente `not-configured`, o `overallStatus` vira `not-configured` quando todos estão nessa condição, e cada sondagem retorna `CureAction=configure_peers`.

O formato esperado de `SOUL_MESH_PEERS` é JSON, por exemplo:

```json
{"N01":"https://n01.example/api","N02":"https://n02.example/api","N03":"https://n03.example/api","N04":"https://n04.example/api","N05":"https://n05.example/api","N06":"https://n06.example/api"}
```

O endpoint final é formado por `SOUL_MESH_ENDPOINT` (padrão `/api/soul-mesh`).

## N07 — Secrets governance

`scripts/pre-sync-check.js` falha antes da sincronização se `SOUL_MESH_TOKEN` ou `REPO_SECRETS_ADMIN_TOKEN` estiver ausente e grava `.diagnostics/pre-sync-status.json`.

`scripts/secret-sync.js` executa a criptografia real com a chave pública de cada repositório e valida o metadata do secret após o PUT.

`scripts/secret-validator.js` verifica presença e frescor sem tentar ler o valor em claro. Igualdade do valor em claro não é exposta pela API do GitHub e não deve ser simulada.

## N07 — E2E certification

`.github/workflows/executor-runner-validation.yml` agora:

1. descobre obrigatoriamente `TestN01ToN07FederatesAcrossN04N05N06`;
2. falha se o teste não existir ou não for descoberto;
3. executa `-count=1 -v` com timeout de 180s;
4. exige marcador real `=== RUN` e `--- PASS` do teste nominal;
5. executa os 30 probes direcionados (15 pares × 2 direções) mesmo quando a etapa nominal falha;
6. arquiva o log e os relatórios para auditoria.

Os 30 probes são provas de protocolo a partir do executor N07. Eles não substituem a prova de tráfego físico originado pelo processo de dois peers quando a topologia não oferece essa instrumentação; o teste federado N01→N07 continua sendo a prova complementar de processo.

## Estado para Frentes 2 e 4

- N02: nova barreira de CI implementada; PR #13 já está merged e deve ser tratado como histórico.
- N07 secrets: código de governança pronto; execução de sincronização permanece bloqueada até configuração real dos dois secrets administrativos.
- N07 peers: dashboard pronto para operar em `not-configured`; URLs reais ainda precisam ser registradas.
- N07 E2E: gate de descoberta e execução real implementado; certificação final depende de o teste ser descoberto e passar em CI.
- N04 e N06 não são tocados por esta Frente.
