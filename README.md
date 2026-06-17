# Service Mesh demo — load balancing de conexões persistentes HTTP/2

Demo ao vivo da talk (Arcos 4 e 7): um serviço Go moderno (HTTP/2, conexão
persistente) **não distribui carga de verdade** porque o `kube-proxy` decide o
destino **por conexão** (L4). A solução é um service mesh (Istio), que opera em
**L7** e balanceia **por request**.

Tema de Copa: o **cliente é o Brasil** e cada request/resposta é um **gol** contra
a seleção adversária (cada pod do servidor é uma seleção, via env `TEAM`). O
cliente mantém um **placar ao vivo**.

| Cenário | Namespace | LB | Resultado |
|--------|-----------|----|-----------|
| Problema (Arco 4) | `no-mesh` | kube-proxy (L4) | Brasil só faz gol contra **UMA** seleção — goleada num pod só |
| Solução (Arco 7) | `mesh` | Envoy / Istio (L7) | gols espalhados **~60/20/20** entre as 3 seleções — por request |

Seleções (pesos fixos dos slides): **Alemanha 60%** (v1, revanche do 7x1),
**Argentina 20%** (v2), **França 20%** (v3) — definidas na env `TEAM` de cada
Deployment em `deploy/base/server.yaml`.

> **Sem TLS em nenhuma camada.** O servidor fala **h2c** (HTTP/2 cleartext) e o
> mTLS automático do Istio fica desligado (`PeerAuthentication: DISABLE`). O
> vilão da talk é a *conexão persistente + L4*, não o TLS — o efeito é idêntico
> em cleartext. (Os slides mencionam "com TLS"; ao vivo é h2c.)

## Pré-requisitos

- **Docker** (testado com 29.x). É o único requisito no host.
- Go, k3d, kubectl e istioctl **não** precisam estar instalados: o Go builda
  dentro de um container e as CLIs são baixadas para `./bin/` pelo `make`.
- Versões pinadas em [`versions.env`](./versions.env).

## Como rodar

```bash
make dev          # build das imagens + cluster k3d + Istio + aplica os 2 namespaces (idempotente)

make demo-no-mesh # streama os logs do cliente: MESMO UUID repetindo  (Arco 4)
make demo-mesh    # streama os logs do cliente: 3 UUIDs ~60/20/20     (Arco 7)

make verify       # checagens: pods do mesh em 2/2 e porta do Service = http2
make clean        # apaga o cluster e ./bin — host sem resíduo
```

Saída esperada do `make demo-no-mesh` (goleada num só):

```
⚽ GOL! Brasil x Alemanha  →  placar: Alemanha 85
⚽ GOL! Brasil x Alemanha  →  placar: Alemanha 86
⚽ GOL! Brasil x Alemanha  →  placar: Alemanha 87
```

Saída esperada do `make demo-mesh` (placar convergindo pra ~60/20/20):

```
⚽ GOL! Brasil x Argentina  →  placar: Alemanha 42  ·  Argentina 22  ·  França 17
⚽ GOL! Brasil x Alemanha   →  placar: Alemanha 43  ·  Argentina 22  ·  França 17
⚽ GOL! Brasil x França     →  placar: Alemanha 43  ·  Argentina 22  ·  França 18
```

A última linha do log já é o placar acumulado. Para isolá-lo:

```bash
KUBECONFIG=./bin/kubeconfig ./bin/kubectl -n mesh logs deploy/go-client --tail=1
```

## Por que funciona

- **Cliente** (`cmd/client`, o "Brasil"): reusa **UMA** conexão HTTP/2 (h2c)
  persistente contra o Service `go-api` (DNS in-cluster, ClusterIP), e cada
  resposta é um gol contra a seleção que respondeu. É o reuso da conexão que
  concentra a carga num pod quando não há mesh. **Não use `kubectl port-forward`**
  para o tráfego da demo — port-forward fixa 1 pod e mascara o comportamento real.
- **Servidor** (`cmd/server`): cada pod representa uma seleção (env `TEAM`) e a
  devolve em `GET /api/v1/health` (com fallback p/ UUID se `TEAM` não for setado).
  São 3 Deployments (`v1/v2/v3`, 1 réplica cada), então cada subset = 1 seleção
  estável.
- **no-mesh**: o `kube-proxy` faz NAT por conexão → a única conexão cai sempre no
  mesmo endpoint → mesmo UUID.
- **mesh**: o sidecar Envoy lê os frames HTTP/2 (L7) e roteia cada request pelos
  pesos do `VirtualService` (60/20/20) sobre os subsets do `DestinationRule`
  (`ROUND_ROBIN`).

### ⚠️ O detalhe que faz ou quebra o Arco 7

A porta do Service **precisa** se chamar `http2` (`appProtocol: http2`). Sem isso o
Envoy trata o tráfego como L4 e o mesh mostraria **um único UUID** — a "solução"
falharia no palco. Veja `deploy/base/server.yaml`.

## Estrutura

```
cmd/server, cmd/client   # Go: servidor h2c + cliente de 1 conexão
Dockerfile.server/client # build multi-stage (host sem Go)
deploy/base              # Deployments v1/v2/v3 + Service + client (kustomize)
deploy/no-mesh           # overlay: namespace no-mesh
deploy/mesh              # overlay: namespace mesh + DestinationRule + VirtualService + PeerAuthentication
Makefile, versions.env   # automação idempotente + versões pinadas
bin/                     # k3d, kubectl, istioctl, kubeconfig (gitignored)
```

## Troubleshooting

- **mesh faz gol só contra uma seleção:** confirme `make verify` (porta `http2`, pods `2/2`).
  Se os pods não estiverem `2/2`, a injeção do sidecar não pegou — `make dev`
  rotula o namespace `mesh` antes de criar os pods; rode `make down && make dev`.
- **trocar k3d por kind:** o Makefile tem `CLUSTER_TOOL=k3d` no topo; as receitas
  `cluster-up`/`images-import`/`clean` precisariam dos equivalentes `kind`.
