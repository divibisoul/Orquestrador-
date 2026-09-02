# SOUL Mesh Secret Bootstrap — N07

## Objetivo

Centralizar o `SOUL_MESH_TOKEN` usado pela Soul Mesh sem registrar o valor em código ou logs.

## Pré-requisitos

O repositório N07 precisa ter dois secrets:

- `SOUL_MESH_TOKEN`: o segredo canônico que será distribuído aos N01–N06.
- `REPO_SECRETS_ADMIN_TOKEN`: token administrativo separado, com permissão de **Actions secrets: write** nos seis repositórios de destino. O `GITHUB_TOKEN` normal do N07 não é suficiente para gravar secrets em outros repositórios.

## Configuração pelo GitHub Web no celular

1. Abra `https://github.com/divibisoul/Orquestrador-/settings/secrets/actions`.
2. Crie/atualize `SOUL_MESH_TOKEN`.
3. Crie/atualize `REPO_SECRETS_ADMIN_TOKEN` com um credencial apropriado e escopo mínimo.
4. Execute o workflow **N07 Soul Mesh Secret Synchronization** em Actions → Run workflow.
5. Confira o artefato `n07-secret-sync-report` e confirme `synced` + `verified: true` para N01–N06.
6. O workflow **N07 Soul Mesh Secret Validation** executa a verificação a cada 6 horas e falha quando algum secret estiver ausente ou sem atualização dentro da janela definida.

## Destinos

- https://github.com/divibisoul/aeternum-core-29/settings/secrets/actions
- https://github.com/divibisoul/Eternium-/settings/secrets/actions
- https://github.com/divibisoul/nexus-aeternum-fusion/settings/secrets/actions
- https://github.com/divibisoul/nextjs-ai-chatbots/settings/secrets/actions
- https://github.com/divibisoul/nextjs-ai-chatbot/settings/secrets/actions
- https://github.com/divibisoul/nextjs-ai-chatbot-2000/settings/secrets/actions

## Limite técnico importante

A API do GitHub não devolve o valor em claro de um secret. Portanto, a validação periódica prova **presença e frescor do secret**, enquanto a sincronização prova que o mesmo valor foi criptografado e enviado pelo N07 para cada destino. A igualdade em claro nunca deve ser obtida por leitura da API.

## Recuperação antifrágil

- `401/403`: revisar a permissão do `REPO_SECRETS_ADMIN_TOKEN` e o acesso aos seis repositórios.
- `404`: revisar nome do repositório e existência do endpoint de secrets.
- `422`: revisar chave pública/criptografia ou o nome do secret.
- Secret `stale`: executar o workflow de sincronização manualmente; não editar o valor em N01–N06 de forma independente.
- Falha do runner: preservar a evidência do job e repetir pelo GitHub Web; não mascarar com `continue-on-error`.
