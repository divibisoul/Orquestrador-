# FASE 0 — N07 TOOLS / CAPABILITIES

## B1 IDENTIDADE
Primário: orquestração, federação e compute.
Secundários: neural, prefrontal, SuperGPU, OctaCore, Mesh, SARA/Jev boundaries.
Papel octacore: processador de coordenação e expansão; não absorve ownership nativo dos outros núcleos.

## B2 PROCESSADORES
| Componente | Path | Responsabilidade | Estado |
|---|---|---|---|
| Engine | orchestrator/engine.go | registry/execução | ativo |
| SuperGPU operations | orchestrator/supergpu_ops.go | compute orchestration | ativo |
| Advanced operations | orchestrator/advanced_ops.go | prefrontal/fusion/federation | ativo |
| Octacore Processor | octacore/ | processor federado | ativo |
| Mesh PeerClient | mesh/ | transporte inter-núcleo | ativo |
| SARAProxy | backend/sara_proxy.go | boundary SARA | ativo quando env |
| JEV client | jev/ | decisão externa | ativo quando env |

## B3 ENDPOINTS
| Método | Path | Estado |
|---|---|---|
| GET | /health | EXECUTABLE |
| GET | /v1/health | EXECUTABLE/BLOCKED_ENV sem N07_APP_TOKEN |
| GET | /v1/capabilities | EXECUTABLE/BLOCKED_ENV sem N07_APP_TOKEN |
| POST | /v1/execute | EXECUTABLE/BLOCKED_ENV sem N07_APP_TOKEN |
| POST | /v1/intent | EXECUTABLE/BLOCKED_ENV sem N07_APP_TOKEN |
| GET | /v1/models | EXECUTABLE |
| POST | /v1/chat/completions | EXECUTABLE conforme boundary OpenAI-compatible |
| GET | /status | BLOCKED_ENV sem token |
| GET | /metrics | EXECUTABLE |
| GET | /identity | EXECUTABLE |
| GET | /topology | EXECUTABLE |
| POST | /api/soul-mesh | EXECUTABLE quando gateway Mesh ativo |
| POST | /execute | BLOCKED_ENV sem bearer |

## B4 FUNÇÕES
| Módulo | Função | Assinatura resumida | Consumidores |
|---|---|---|---|
| Engine | Register | (operation,handler) => error | startup |
| Engine | Execute | (ctx,operation,payload,metadata) => Result | HTTP/internal |
| Engine | Operations | () => []string | capabilities |
| Engine | Health | () => map | health |
| SuperGPU | RegisterSuperGPUOperations | (Engine) => error | startup |
| SARAProxy | Cycle/Audit/Regenerate/State/Capabilities | HTTP boundary methods | N07 |
| Octacore | RegisterOperations | (Engine,Processor) => error | startup |

## B5/B6 EVENTOS
N07 publica Vagus via SARAProxy quando configurado. Soul Mesh trata request/response/event. Lista exaustiva de eventos internos não foi enumerada: PENDING.

## B7 EXTERNOS
SARA service, Jev/TypeSafe, Supabase/Web3 storage, peer URLs, Mesh secrets/HMAC.

## B8 INTER-NÚCLEO
Soul Mesh é transporte canônico. N07 chama peers pelo client existente; Octacore usa peer transport; SARA é boundary HTTP; Vagus é evento e não Mesh.

## B9 ADORMECIDAS
| Ferramenta | Precisa de | Estado |
|---|---|---|
| SARA ops | SARA_SERVICE_URL + SARA_SERVICE_TOKEN | BLOCKED_ENV |
| Jev | JEV_API_KEY | BLOCKED_ENV |
| RGO Trinity | PR #54 + serviços reais | BRANCH_ONLY |
| external storage | Supabase/Web3 env | BLOCKED_ENV |
| remote Mesh | peer URLs/auth | BLOCKED_ENV |

## B10 EXECUTÁVEIS
neural.forward, neural.learn, compute.execute, cognitive.execute, supergpu.*, octacore.* e Mesh gateway possuem handlers no MAIN. Isso é execução de código; LIVE exige ambiente + transação observada.

## B11 EXPANSÃO
| Ao conectar | Ganha | Perde | Neutro |
|---|---|---|---|
| N01 | runtime/Mesh/Horta | nenhuma | orchestration |
| N02 | cognitive/generation | nenhuma | routing |
| N03 | perception/audio | nenhuma | compute |
| N04 | tools/docs | nenhuma | routing |
| N05 | inference | nenhuma | policy |
| N06 | context/session | nenhuma | orchestration |
| SARA | regeneration/governance | nenhuma | N07 authority |

Inventário recursivo total de exports/eventos permanece PENDING onde a ferramenta não expõe árvore/checkout integral.
