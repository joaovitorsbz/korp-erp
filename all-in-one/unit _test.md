# Cenários de Teste — Projeto Korp

Este documento reúne os testes que validam o projeto de ponta a ponta:
serviço, proxy reverso, rede Docker, monitoramento e a automação via
Ansible. Cada cenário tem objetivo, pré-condição, passo a passo e o
resultado esperado.

Pré-requisito geral: stack no ar (manualmente com `docker compose up -d`
ou via `ansible-playbook -i inventory.ini playbook.yml -K`).

---

## 1. Serviço HTTP (`http-server-projeto-korp`)

### 1.1 — Endpoint principal retorna o JSON esperado

**Objetivo:** confirmar que `/projeto-korp` responde com a estrutura correta.

```bash
curl -s http://localhost/projeto-korp
```

**Esperado:**
```json
{"nome": "Projeto Korp", "horario": "2026-08-08T12:00:00Z"}
```

### 1.2 — Horário é dinâmico a cada requisição

**Objetivo:** confirmar que o horário não é fixo/cacheado.

```bash
curl -s http://localhost/projeto-korp; sleep 2; curl -s http://localhost/projeto-korp
```

**Esperado:** os dois valores de `horario` são diferentes (ou, se a
requisição cair no mesmo segundo, rode de novo com um `sleep` maior).

### 1.3 — Container do app não expõe porta ao host

**Objetivo:** validar o requisito "não deve expor portas diretamente ao host".

```bash
docker port http-server-projeto-korp
```

**Esperado:** saída vazia (nenhuma porta publicada). O serviço só deve
ser acessível através do NGINX, na porta 80.

---

## 2. Proxy reverso (NGINX)

### 2.1 — Proxy encaminha corretamente para o serviço

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://localhost/projeto-korp
```

**Esperado:** `200`

### 2.2 — Página inicial (landing page) responde

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://localhost/
```

**Esperado:** `200`, com HTML contendo o botão para `/projeto-korp`.

### 2.3 — `/metrics` acessível via proxy

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://localhost/metrics
```

**Esperado:** `200`, corpo no formato de métricas do Prometheus
(`# HELP`, `# TYPE`, etc.).

---

## 3. Build e rede Docker

### 3.1 — Build da imagem sem erros

```bash
docker compose build http-server-projeto-korp
```

**Esperado:** build finaliza com sucesso (exit code 0).

### 3.2 — Rede `korp-net` existe e é do tipo bridge

```bash
docker network inspect korp-net --format '{{.Driver}}'
```

**Esperado:** `bridge`

### 3.3 — Todos os containers do projeto estão na mesma rede

```bash
docker network inspect korp-net --format '{{range .Containers}}{{.Name}} {{end}}'
```

**Esperado:** lista contém `http-server-projeto-korp`, `nginx`,
`nginx-exporter`, `prometheus` e `grafana`.

---

## 4. Monitoramento (Prometheus)

### 4.1 — Target do serviço está `UP`

```bash
curl -s http://localhost:9090/api/v1/query --data-urlencode 'query=up{job="http-server-projeto-korp"}'
```

**Esperado:** `"value":[...,"1"]`

### 4.2 — Volume de requisições é contabilizado

```bash
for i in $(seq 1 5); do curl -s http://localhost/projeto-korp >/dev/null; done
curl -s http://localhost:9090/api/v1/query --data-urlencode 'query=sum(http_requests_total{job="http-server-projeto-korp"})'
```

**Esperado:** o valor aumenta a cada rodada do laço `for` (mínimo +5).

### 4.3 — Disponibilidade do NGINX é monitorada de forma independente

**Objetivo:** validar que a disponibilidade do *proxy* (não só do backend)
é medida — ver cenário de falha 6.2 para o teste completo.

```bash
curl -s http://localhost:9090/api/v1/query --data-urlencode 'query=nginx_up'
```

**Esperado:** `"value":[...,"1"]` com o NGINX saudável.

> **Atenção:** não use `up{job="nginx"}` para checar isso — essa métrica
> só indica que o *exporter* respondeu, não que o NGINX em si está de pé
> (ver seção 6.2 para o motivo).

---

## 5. Dashboard (Grafana)

### 5.1 — Datasource e dashboard provisionados automaticamente

```bash
curl -s -u admin:admin http://localhost:3000/api/datasources
curl -s -u admin:admin http://localhost:3000/api/search?query=korp
```

**Esperado:** datasource `Prometheus` (uid `prometheus-korp`) e dashboard
`http-server-projeto-korp` aparecem **sem nenhum clique manual** — só de
a stack subir.

### 5.2 — Painéis carregam sem erro de datasource

Acesse `http://localhost:3000/d/http-server-projeto-korp` no navegador,
logue com `admin`/`admin` e confirme visualmente:

- **Disponibilidade da aplicação (backend)** — verde
- **Disponibilidade do NGINX (proxy / entrada externa)** — verde
- **Requisições por segundo** — gráfico com dados após gerar tráfego
- **Total de requisições** — número crescente
- **Requisições por status** — pizza com fatia `200`

---

## 6. Cenários de falha (o que interessa de verdade)

### 6.1 — Backend cai, proxy continua de pé

**Objetivo:** confirmar que o dashboard detecta corretamente quando o
*serviço* (não o proxy) fica indisponível.

```bash
docker stop http-server-projeto-korp
sleep 10
curl -s http://localhost:9090/api/v1/query --data-urlencode 'query=up{job="http-server-projeto-korp"}'
```

**Esperado:** valor cai para `0` em até ~5-10s (intervalo de scrape do
Prometheus é 5s). No Grafana, o painel "Disponibilidade da aplicação
(backend)" fica vermelho.

**Restaurar:** `docker start http-server-projeto-korp`

### 6.2 — NGINX cai, backend continua saudável (cenário crítico)

**Objetivo:** este é o teste mais importante do projeto — validar que a
queda do *mediador* (ponto único de entrada externo) é detectada, mesmo
que o serviço por trás dele esteja 100% saudável.

```bash
docker stop nginx
sleep 10
curl -s http://localhost:9090/api/v1/query --data-urlencode 'query=nginx_up'
curl -s http://localhost:9090/api/v1/query --data-urlencode 'query=up{job="http-server-projeto-korp"}'
```

**Esperado:**
- `nginx_up` → `0` (o proxy caiu)
- `up{job="http-server-projeto-korp"}` → `1` (o backend continua de pé —
  ele só não está mais acessível de fora, porque o único caminho público
  é através do NGINX)

No Grafana: painel "Disponibilidade do NGINX" fica vermelho, painel
"Disponibilidade da aplicação (backend)" continua verde. Essa diferença
visual é o ponto principal do teste.

Tentativa de acesso externo enquanto o NGINX está parado:

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://localhost/projeto-korp
```

**Esperado:** erro de conexão (`curl: (7) Failed to connect`) — a
aplicação está de pé, mas ninguém de fora consegue alcançá-la.

**Restaurar:** `docker start nginx`

### 6.3 — Porta 80 do host ocupada por outro processo

**Objetivo:** validar o mecanismo de porta alternativa via variável de
ambiente, usado quando a porta 80 já está em uso na máquina de destino.

```bash
KORP_HTTP_PORT=8081 docker compose up -d --force-recreate nginx
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8081/projeto-korp
```

Ou, via Ansible:

```bash
ansible-playbook -i inventory.ini playbook.yml -K --extra-vars "korp_http_port=8081"
```

**Esperado:** `200` na porta alternativa, sem precisar editar o
`docker-compose.yml`.

### 6.4 — Alterar arquivo de configuração montado exige recriar o container

**Objetivo:** demonstrar por que o playbook usa `--force-recreate` — um
container já rodando não relê sozinho um arquivo montado (ex:
`prometheus.yml`, conf do NGINX) que mudou no disco.

```bash
# edite algo em monitoring/prometheus/prometheus.yml, depois:
docker compose up -d prometheus            # sem --force-recreate
docker exec prometheus cat /etc/prometheus/prometheus.yml   # ainda mostra o conteudo antigo, se o container ja existia
docker compose up -d --force-recreate prometheus
docker exec prometheus cat /etc/prometheus/prometheus.yml   # agora reflete a mudanca
```

**Esperado:** só o `--force-recreate` garante que a mudança é aplicada.
É por isso que o playbook Ansible sempre usa essa flag.

---

## 7. Automação (Ansible)

### 7.1 — Provisionamento do zero com um único comando

**Pré-condição:** ambiente limpo (`docker compose down --rmi all
--volumes` e `docker network rm korp-net`, se existir).

```bash
ansible-playbook -i inventory.ini playbook.yml -K
```

**Esperado:** playbook roda sem falhas até o fim, e a última task
(`validate : Exibir resposta do servico no console`) imprime o JSON do
`/projeto-korp` no terminal — prova de que o ambiente foi validado
automaticamente, sem intervenção manual.

### 7.2 — Playbook é idempotente (rodar duas vezes não quebra nada)

```bash
ansible-playbook -i inventory.ini playbook.yml -K
ansible-playbook -i inventory.ini playbook.yml -K
```

**Esperado:** a segunda execução termina igualmente bem-sucedida
(containers recriados de novo, sem erro), e o teste final continua
retornando `200`.

### 7.3 — Compatibilidade entre distribuições

**Objetivo:** validar a role `docker` nas duas famílias de Linux suportadas.

| Distro                | Gerenciador de pacote | Resultado esperado |
|------------------------|------------------------|---------------------|
| Ubuntu 22.04/24.04     | `apt`                  | Docker + `docker-compose-v2` instalados via apt |
| Amazon Linux 2023      | `dnf`                  | Docker via `dnf`, plugin `docker compose` baixado do GitHub releases |

Rodar o mesmo comando do item 7.1 em cada distro e confirmar que ambas
chegam ao mesmo resultado final (JSON exibido no console).
