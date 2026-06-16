# Roteiro — Load Balancing com Service Mesh em Go
> GolangSP — Talk proposta por Vinicius Lopes

---

## Contexto Geral

Especificação completa da apresentação. Registra todas as decisões de framework, narrativa e estrutura já tomadas.

**Artigo de referência:** https://dev.to/visepol/load-balancing-grpc-traffic-with-istio-1k49  
O artigo cobre o problema e a solução usando gRPC + Istio. A talk expande esse conteúdo correlacionando com Go padrão (HTTP/2 sobre TLS) e tornando o problema mais amplo para a audiência Go.

---

## 1. Briefing

| Campo | Valor |
|---|---|
| Evento | GolangSP — meetup da comunidade Go de São Paulo |
| Formato | Palestra técnica com demo ao vivo |
| Tempo total | ~40 minutos |
| Breakdown estimado | 30 min talk + demo / 10 min Q&A |
| Idioma | Português |

**Perfil da audiência:**
- Engenheiros sênior e staff
- Familiaridade com Go e microserviços
- Não necessariamente familiarizados com service mesh ou Istio
- Alta tolerância técnica, baixa tolerância para rodeio ou conteúdo óbvio
- Respondem a evidência, não a teoria

**Take-home message:**
> Seu serviço Go com TLS provavelmente não está distribuindo carga de verdade — e a solução não está no seu código.

---

## 2. Framework de Comunicação

### Âncora: SCQA (global)

- **Situation:** Go usa HTTP/2 por padrão. Microserviços Go em produção são comuns.
- **Complication:** Infraestrutura L4 não enxerga requests dentro de conexões persistentes — toda a carga vai para o mesmo pod.
- **Question:** Como distribuir carga de verdade sem mudar o código da aplicação?
- **Answer:** Service mesh opera em L7 e resolve na camada certa, de forma transparente para a aplicação.

### Modelo narrativo: Nancy Duarte — Resonate (Sparkline)

A apresentação alterna entre dois estados ao longo do tempo:

- **What Is** — realidade atual, o problema que a audiência provavelmente já viveu sem saber a causa
- **What Could Be** — o mundo com service mesh, onde a infra resolve o que é responsabilidade dela

### Papel do apresentador

O apresentador é o **mentor**, não o herói. A audiência é quem vai implementar, operar e defender essa solução nos seus times.

### STAR Moment

Demo ao vivo onde a audiência vê, em tempo real, que sem Istio todos os responses vêm do mesmo UUID de servidor — e com Istio, os UUIDs variam. É evidência, não argumento.

---

## 3. Arco Macro

```
[What Is]       Go + HTTP/2 + TLS é o padrão. Microserviços Go em prod são realidade.
      ↓
[Complication]  L4 não enxerga dentro da conexão. Carga vai para um único pod.
      ↓
[STAR]          Demo ao vivo: UUID prova o problema em produção simulada.
      ↓
[What Could Be] Service mesh como camada L7. Sidecar proxy. Controle sem tocar no código.
      ↓
[Answer]        Com Istio: demo mostra distribuição real. Algoritmos, routing, DestinationRules.
      ↓
[New Bliss]     A infra cuida do que é da infra. Seu código Go continua o mesmo.
```

---

## 4. Decisões Registradas

| Decisão | Escolha | Motivo |
|---|---|---|
| Framework de comunicação | SCQA global | Evita repetição, mantém arco único de tensão |
| Modelo narrativo | Duarte Sparkline | Cria ritmo emocional para audiência técnica sênior |
| Formato de evidência | Demo ao vivo com UUIDs | Evidência > argumento para audiência sênior |
| Slides | Construídos após roteiro | Slide é mídia, não roteiro |
| Papel do apresentador | Mentor | Audiência é o agente de mudança |
| Título da apresentação | Conexões Persistentes em Go — Controlando o Tráfego com Service Mesh | Nomeia problema e solução; assentamento, não gancho |
| Demo — algoritmo | ROUND_ROBIN com pesos 60/20/20 | Efeito visível ao vivo; LEAST_CONN difícil de demonstrar |
| Narrativa L4/L7 | "L4 enxerga conexão. L7 enxerga request." | "Pacote" é terminologia de L3 — causa ruído com audiência sênior |

---

## 5. Estrutura de Arcos

| # | Título | Posição | Slides | Duração |
|---|---|---|---|---|
| 1 | Conexões Persistentes em Go — Controlando o Tráfego com Service Mesh | What Is | 5 | ~6 min |
| 2 | Conexões persistentes: a vantagem que vira armadilha | What Is → Complication | 4 | ~6 min |
| 3 | Por que o proxy não consegue ajudar | Complication (escalada) | 3 | ~5 min |
| 4 | Demo — vendo carga ir para um único pod em tempo real | STAR | 2 | ~5 min |
| 5 | O que muda quando você opera em L7 | What Could Be | 4 | ~6 min |
| 6 | Sidecar proxy — como L7 vira realidade sem mudar o código | Answer | 2 | ~4 min |
| 7 | Demo — vendo a carga distribuída em tempo real | Answer → New Bliss | 3 | ~6 min |
| 8 | A infra cuida do que é da infra | New Bliss → Call to Action | 2 | ~4 min |

**Total: 25 slides / ~42 min** *(os 2 slides de fundamentos do Arco 1 somam ~+2 min ao tempo de talk)*

---

## 6. Descrição dos Slides

---

### Arco 1 — What Is (~6 min)
**`Conexões Persistentes em Go — Controlando o Tráfego com Service Mesh`**
*Resumo: a web nasceu de duas peças simples — HTTP + HTML — e foi essa simplicidade (stateless, texto legível, request/response) que a fez escalar. Mas abrir conexão custa: o handshake. Go usa HTTP/2 por padrão, que reaproveita uma única conexão TCP para evitar esse custo — menos overhead, melhor performance. É o stack que a audiência já opera e o ponto de partida da talk.*

---

#### Slide 1.1 — Título + apresentador

| Campo | Decisão |
|---|---|
| Conteúdo | Título da talk + nome do apresentador + evento |
| Formato visual | Texto + logo GolangSP + logo Gopher |
| Imagem/ícone | Ícone sutil de rede/malha no fundo, baixa opacidade |
| Animação | Estático, aparece de uma vez |
| Speaker note | Apresentação pessoal breve + setup da narrativa da talk |
| Transição | Pergunta de engajamento — "HTTP/2, todo mundo usando?" |

---

#### Slide 1.2 — HTTP + HTML: a web simples

| Campo | Decisão |
|---|---|
| Conteúdo | História de HTTP + HTML e por que funcionavam tão bem — a simplicidade que fez a web escalar |
| Formato visual | Par HTTP + HTML no topo + grade de 4 propriedades |
| Imagem/ícone | Ícones Phosphor (HTTP, HTML) e por propriedade (stateless, texto, request/response, desacoplado) |
| Animação | De uma vez — fadeIn |
| Speaker note | A web nasceu de duas peças de texto; stateless/legível/request-response foi o que permitiu escalar. Fixar "stateless": cada request é independente |
| Transição | Virada — "mas abrir uma conexão tem um custo, e esse custo tem nome: handshake" |

**Estrutura:**
```
HTTP — como pedir e transportar   +   HTML — o que é transportado

Stateless · Texto legível · Request/Response · Desacoplado
```

---

#### Slide 1.3 — Toda conexão nasce e morre

| Campo | Decisão |
|---|---|
| Conteúdo | Ciclo de vida de uma conexão — handshake (SYN → SYN-ACK → ACK) abre, conteúdo passa (GET → 200), e a conexão morre (FIN). O ponto: recebeu o conteúdo, a conexão acaba |
| Formato visual | Colunas Cliente / Servidor + 6 setas direcionais (3 handshake com badge, divisor "conexão aberta", entrega, FIN em vermelho) + punchline "a conexão morre" |
| Imagem/ícone | Ícones Cliente (desktop) e Servidor; setas desenhadas em CSS; FIN destacado em vermelho |
| Animação | Build beat a beat — cada passo revelado em sequência até a morte da conexão; reinicia a cada navegação (`.slide--active`) |
| Speaker note | Handshake abre, conteúdo passa, FIN encerra. Recebeu o conteúdo → a conexão morre → o próximo request começa tudo de novo. Esse custo repetido é a dor que a evolução resolve |
| Transição | Antecipação — "pagar o handshake a cada request é caro. Foi isso que a evolução veio resolver" |

**Estrutura:**
```
Cliente                    Servidor
   ── SYN ───────────────────▶     (1) "abrir conexão"
   ◀─────────────── SYN-ACK ──     (2) "conexão aberta"
   ── ACK ───────────────────▶     (3) "confirmado"
   ───────── conexão aberta ─────────
   ── GET / ──────────────────▶     "manda o conteúdo"
   ◀──────────── 200 + dados ──     "aqui está"
   ── FIN ───────────────────▶     ✕ conexão encerrada

   Recebeu o conteúdo? A conexão morre.
   Próximo request: tudo de novo.
```

---

#### Slide 1.4 — O stack Go

| Campo | Decisão |
|---|---|
| Conteúdo | Evolução HTTP/1.1 → HTTP/2 — o problema que a multiplexação resolve |
| Formato visual | Diagrama progressivo comparativo |
| Imagem/ícone | Nenhum além do diagrama |
| Animação | Build em 3 etapas: HTTP/1.1 → problema destacado → HTTP/2 |
| Speaker note | Custo concreto do HTTP/1.1 — handshakes, overhead, conexões simultâneas |
| Transição | Virada — multiplexação é vantagem, mas a conexão persistente tem uma propriedade que vai importar |

---

#### Slide 1.5 — Go em produção

| Campo | Decisão |
|---|---|
| Conteúdo | Arquitetura típica — cliente → ingress → múltiplos pods Go |
| Formato visual | Diagrama estilo Kubernetes — ícones K8s, ingress nginx, namespace |
| Imagem/ícone | Logo do Go nos pods |
| Animação | Build por camada: pods → ingress → cliente |
| Speaker note | O ambiente parece saudável — mas tem algo acontecendo que o diagrama não mostra |
| Transição | Pergunta retórica — "qual pod está respondendo mais requests agora?" |

---

### Arco 2 — What Is → Complication (~6 min)
**`Conexões persistentes: a vantagem que vira armadilha`**
*Resumo: A audiência sai sabendo que seu cluster provavelmente está distribuindo conexões, não requests — e que isso significa um pod absorvendo carga desproporcional sem nenhum alarme óbvio.*

---

#### Slide 2.1 — Multiplexação revisitada

| Campo | Decisão |
|---|---|
| Conteúdo | Propriedade da conexão persistente — handshake uma vez, N requests sobre a mesma conexão |
| Formato visual | Linha do tempo — handshake + requests sequenciais |
| Imagem/ícone | Nenhum além da linha do tempo |
| Animação | Build em etapas: handshake → requests chegando um a um |
| Speaker note | Custo evitado pelo HTTP/2 + comportamento padrão do runtime Go com TLS |
| Transição | Pergunta — "o que acontece com os outros pods quando a carga aumenta?" |

---

#### Slide 2.2 — O que o ingress enxerga

| Campo | Decisão |
|---|---|
| Conteúdo | Por que o L4 não consegue agir — distribui conexões, não requests |
| Formato visual | Diagrama com duas perspectivas: ingress (conexão opaca) vs. interior (requests num único pod) |
| Imagem/ícone | Nenhum além do diagrama |
| Animação | De uma vez — duas perspectivas juntas |
| Speaker note | Analogia da rodovia — ingress vê um carro, não sabe quantos passageiros estão dentro |
| Transição | Consequência — "o que acontece em produção quando a carga aumenta?" |

---

#### Slide 2.3 — Um pod absorve tudo

| Campo | Decisão |
|---|---|
| Conteúdo | Contraste visual — pod sobrecarregado vs. pods ociosos |
| Formato visual | Comparativo de barras — um no limite, dois em zero |
| Imagem/ícone | Ícones K8s nos pods — consistência com Arco 1 |
| Animação | De uma vez — contraste imediato |
| Speaker note | CPU média de 33% parece saudável — a média esconde um pod a 100% e dois em zero |
| Transição | Provocação — escalar horizontalmente não resolve; a conexão já está estabelecida |

---

#### Slide 2.4 — A pergunta que você não consegue responder

| Campo | Decisão |
|---|---|
| Conteúdo | Três perguntas sem resposta com o stack atual |
| Formato visual | Texto puro — três perguntas em tipografia grande, uma por linha |
| Imagem/ícone | Nenhum |
| Animação | De uma vez — três perguntas juntas |
| Speaker note | Ponte para a demo — essas perguntas terão resposta visível em tempo real |
| Transição | Pergunta à audiência — "alguém já ouviu falar de L4 e L7 do modelo OSI?" |

---

### Arco 3 — Complication (escalada) (~5 min)
**`Por que o proxy não consegue ajudar`**
*Resumo: A audiência sai com o modelo mental de que a pilha OSI não é teoria acadêmica — é o motivo pelo qual o ingress nginx se comporta de forma contraintuitiva em produção. Operando em L4, qualquer proxy enxerga fluxo TCP, não requests HTTP: a conexão é opaca por design. O problema não é configuração, não é produto — é a camada.*

---

#### Slide 3.1 — A pilha OSI

| Campo | Decisão |
|---|---|
| Conteúdo | Pilha OSI com foco em L4 e L7 — demais camadas em segundo plano |
| Formato visual | Diagrama vertical — 7 camadas, L4 e L7 com destaque visual |
| Imagem/ícone | Nenhum além do diagrama |
| Animação | De uma vez — destaque já aplicado |
| Speaker note | Recorte intencional (L1-L3, L5-L6 ignorados) + ponte ingress em L4 vs. Istio em L7 |
| Transição | Afirmação — "L4 enxerga fluxo TCP. Não enxerga HTTP, requests ou headers." |

---

#### Slide 3.2 — O que o ingress nginx enxerga

| Campo | Decisão |
|---|---|
| Conteúdo | O que o ingress nginx enxerga em L4 — conexão TCP, bytes opacos, sem acesso ao HTTP |
| Formato visual | Representação de pacote TCP — bloco opaco como payload |
| Imagem/ícone | Nenhum além do bloco opaco |
| Animação | De uma vez — pacote completo aparece junto |
| Speaker note | Analogia dos Correios — carteiro entrega pelo endereço, não abre o envelope |
| Transição | Generalização — nginx, HAProxy, qualquer L4 tem o mesmo limite; a camada define o que é possível |

---

#### Slide 3.3 — Não é o nginx, é a camada

| Campo | Decisão |
|---|---|
| Conteúdo | Conclusão em uma frase — "não é o nginx, não é a configuração, é a camada" |
| Formato visual | Texto puro — frase centralizada em tipografia grande |
| Imagem/ícone | Nenhum |
| Animação | De uma vez — conclusão entregue sem fragmentação |
| Speaker note | Ponte para a solução — a solução não é corrigir o proxy, é operar em outra camada |
| Transição | Ação — "vamos ver o problema ao vivo. UUID mostrando qual pod responde cada request." |

---

### Arco 4 — STAR (~5 min)
**`Demo — vendo carga ir para um único pod em tempo real`**
*Resumo: A audiência vê em tempo real que todos os requests vão para o mesmo UUID — o mesmo pod respondendo enquanto os outros ficam ociosos. O problema deixa de ser teoria e vira evidência.*

---

#### Slide 4.1 — Setup da demo

| Campo | Decisão |
|---|---|
| Conteúdo | Arquitetura do ambiente — cluster local, três pods Go, ingress nginx, cliente em loop |
| Formato visual | Screenshot real do kubectl get pods |
| Imagem/ícone | Nenhum além do screenshot |
| Animação | De uma vez |
| Speaker note | Contrato da demo — cada pod responde com UUID próprio, o padrão fica claro rapidamente |
| Transição | Ação direta — "ambiente pronto. Vou abrir o terminal agora." |

---

#### Slide 4.2 — Terminal ao vivo

| Campo | Decisão |
|---|---|
| Conteúdo | Terminal ao vivo — requests em loop, UUID repetindo, mesmo pod respondendo |
| Formato visual | Terminal em tela cheia — fonte grande, fundo escuro |
| Imagem/ícone | Nenhum |
| Animação | Sem animação — o terminal ao vivo é o conteúdo |
| Speaker note | Narração direta — "reparem no UUID, é sempre o mesmo. Os outros dois pods estão ociosos agora mesmo." |
| Transição | Diagnóstico — "a solução não está no código, não está no nginx — está em operar em outra camada" |

---

### Arco 5 — What Could Be (~6 min)
**`O que muda quando você opera em L7`**
*Resumo: A audiência entende por que L7 é a camada certa para esse problema — e que service mesh não é over-engineering, é a solução arquiteturalmente adequada para o que foi demonstrado.*

---

#### Slide 5.1 — L4 vs L7

| Campo | Decisão |
|---|---|
| Conteúdo | "L4 enxerga conexão. L7 enxerga request." |
| Formato visual | Texto puro — duas frases em tipografia grande |
| Imagem/ícone | Nenhum |
| Animação | De uma vez |
| Speaker note | "Você só pode agir sobre o que você enxerga — a camada define o que é possível, não a configuração" |
| Transição | Afirmação — "existe uma camada de infraestrutura projetada para operar em L7" |

---

#### Slide 5.2 — O que é service mesh

| Campo | Decisão |
|---|---|
| Conteúdo | Dois blocos — Control Plane (Istiod) e Data Plane (sidecar Envoy) com bullets técnicos |
| Formato visual | Bullets em dois blocos estruturados |
| Imagem/ícone | Logo Istio acima do Control Plane, logo Envoy acima do Data Plane |
| Animação | De uma vez — dois blocos juntos |
| Speaker note | Transparência para o código Go (zero mudança) + fluxo de controle: Istiod distribui regras, sidecars executam localmente |
| Transição | Zoom — "como o sidecar se posiciona dentro do pod e intercepta o tráfego sem que a aplicação saiba" |

**Estrutura dos bullets:**
```
Control Plane (Istiod)
• Interage com a API do Kubernetes
• Injeta o sidecar proxy em cada pod
• Distribui as regras de roteamento

Data Plane (sidecar Envoy)
• Roda junto com a aplicação no mesmo pod
• Intercepta todo o tráfego de entrada e saída
• Opera em L7 — enxerga o request
• Reporta métricas ao control plane
```

---

#### Slide 5.3 — Arquitetura de sidecar

| Campo | Decisão |
|---|---|
| Conteúdo | Arquitetura interna do pod — Go + Envoy, interceptação via iptables, aplicação não percebe |
| Formato visual | Diagrama do pod com dois containers e setas de fluxo |
| Imagem/ícone | Logo Go no container da aplicação, logo Envoy no sidecar |
| Animação | Build em etapas: pod → containers → setas de tráfego |
| Speaker note | Mecanismo iptables + zero mudança no código Go + capacidades do Envoy em L7 (load balancing, circuit breaking, retry, mTLS) |
| Transição | Sequência — "adotar isso não é complexidade, é devolver responsabilidade para a camada certa" |

---

#### Slide 5.4 — Por que isso não é over-engineering

| Campo | Decisão |
|---|---|
| Conteúdo | Trade-off explícito — "O que você paga" vs. "O que você ganha" + conclusão na separação de responsabilidades |
| Formato visual | Duas colunas com bullets |
| Imagem/ícone | Nenhum |
| Animação | De uma vez |
| Speaker note | Comparação com alternativas — você paga o custo de qualquer forma, só que no código em vez da infra |
| Transição | Ação — "agora vamos ver o Istio configurado e a demo com distribuição real de carga" |

**Estrutura das colunas:**
```
O que você paga                     O que você ganha
• Sidecar por pod                   • Load balancing por request
• Control plane no cluster          • Observabilidade em L7
• Regras de iptables gerenciadas    • mTLS automático
• Curva de aprendizado do Istio     • Circuit breaking
                                    • Retry
                                    — sem uma linha de código Go
```

---

### Arco 6 — Answer (~4 min)
**`Sidecar proxy — como L7 vira realidade sem mudar o código`**
*Resumo: A audiência entende como o sidecar proxy intercepta tráfego em L7 sem que a aplicação Go saiba — e por que esse design é a resposta arquitetural correta para o problema demonstrado na demo.*

---

#### Slide 6.1 — Algoritmos de balanceamento em L7

| Campo | Decisão |
|---|---|
| Conteúdo | Três algoritmos — ROUND_ROBIN, LEAST_CONN, RANDOM — com LEAST_CONN em destaque conectado à demo. Demo usará ROUND_ROBIN com pesos. |
| Formato visual | Cards lado a lado — LEAST_CONN maior com borda destacada |
| Imagem/ícone | Indicador visual no LEAST_CONN conectando ao desequilíbrio do Arco 4 |
| Animação | De uma vez |
| Speaker note | "Para usar LEAST_CONN no Istio, você declara no campo loadBalancer da DestinationRule. A demo usará ROUND_ROBIN com pesos 60/20/20 para tornar o efeito visível ao vivo." |
| Transição | Antecipação — "o YAML exato que será aplicado na demo, com pesos diferentes para cada pod" |

---

#### Slide 6.2 — DestinationRule + VirtualService

| Campo | Decisão |
|---|---|
| Conteúdo | YAML mínimo funcional — DestinationRule + VirtualService com pesos 60/20/20 |
| Formato visual | Bloco único de código com `---` separando os dois objetos |
| Imagem/ícone | Nenhum |
| Animação | De uma vez |
| Speaker note | "DestinationRule define subsets com ROUND_ROBIN. VirtualService distribui 60/20/20. kubectl apply sem restart de pod." |
| Transição | Antecipação — "três pods, pesos 60/20/20, ROUND_ROBIN em L7. O mesmo ambiente da primeira demo — com uma configuração diferente. Vamos ver." |

---

### Arco 7 — Answer → New Bliss (~6 min)
**`Demo — vendo a carga distribuída em tempo real`**
*Resumo: A demo fecha o arco narrativo aberto no Arco 4 — o problema foi visto ao vivo, a solução é vista ao vivo. A audiência sai com o contraste completo e a confiança de que a solução é real e acessível.*

---

#### Slide 7.1 — Setup da demo

| Campo | Decisão |
|---|---|
| Conteúdo | Mesmo ambiente do Arco 4 — agora com Istio e YAML aplicados |
| Formato visual | Screenshot kubectl get pods — 2/2 containers por pod |
| Imagem/ícone | Anotação destacando coluna READY mostrando 2/2 |
| Animação | De uma vez |
| Speaker note | "Mesmo cluster, mesmos pods, mesmo cliente. Única diferença: Istio. Nenhuma linha de código Go mudou." |
| Transição | Ação direta — "ambiente pronto. Vou abrir o terminal agora." |

---

#### Slide 7.2 — Terminal ao vivo

| Campo | Decisão |
|---|---|
| Conteúdo | Terminal ao vivo — três UUIDs alternando nas proporções 60/20/20 |
| Formato visual | Terminal em tela cheia — espelho do Slide 4.2 |
| Imagem/ícone | Nenhum |
| Animação | Sem animação — terminal ao vivo é o conteúdo |
| Speaker note | "Reparem nos UUIDs — agora são três alternando. Pod A com 60%, os outros dois dividem os 40%." |
| Transição | Afirmação — "mesmo ambiente, mesmo código Go. Uma camada diferente — carga distribuída." |

---

#### Slide 7.3 — O contraste

| Campo | Decisão |
|---|---|
| Conteúdo | Contraste — sem Istio (um pod a 100%) vs. com Istio (distribuição 60/20/20) |
| Formato visual | Comparativo de barras — dois estados lado a lado |
| Imagem/ícone | Nenhum |
| Animação | Barras crescendo ao entrar no slide — uma sobe a 100%, depois as três crescem em 60/20/20 |
| Speaker note | "Isso não é só load balancing — observabilidade, mTLS, circuit breaking, retry. Tudo via YAML, zero código." |
| Transição | Fechamento — "o problema estava na camada. A solução estava na camada. O código Go não mudou." |

---

### Arco 8 — New Bliss → Call to Action (~4 min)
**`A infra cuida do que é da infra`**
*Resumo: A audiência sai com um modelo mental claro: o código Go não muda, a infra evolui. E com a confiança de que essa evolução é incremental — não precisa de big bang para começar.*

---

#### Slide 8.1 — A tese

| Campo | Decisão |
|---|---|
| Conteúdo | Separação de responsabilidades como princípio — código cuida de lógica, infra cuida de tráfego |
| Formato visual | Texto puro — frase centralizada em tipografia grande |
| Imagem/ícone | Nenhum |
| Animação | De uma vez |
| Speaker note | "A próxima vez que alguém propuser retry no código — a pergunta é: isso é responsabilidade do código ou da infra?" |
| Transição | Entrega — "repositório, artigo e contato no próximo slide. Levem para casa." |

---

#### Slide 8.2 — Referências + contato

| Campo | Decisão |
|---|---|
| Conteúdo | QR code para o artigo + @ GitHub + @ LinkedIn |
| Formato visual | QR code centralizado e grande, handles abaixo |
| Imagem/ícone | Ícone do GitHub e ícone do LinkedIn ao lado de cada handle |
| Animação | De uma vez |
| Speaker note | "Slide fica no ar durante as perguntas. Quem quiser fotografar o QR code, fique à vontade. Perguntas?" |
| Transição | Nenhuma — último slide, fica no ar durante o Q&A |

---

## Referências

- Artigo base: https://dev.to/visepol/load-balancing-grpc-traffic-with-istio-1k49
- Nancy Duarte — Resonate (livro)
- Minto Pyramid Principle — SCQA
