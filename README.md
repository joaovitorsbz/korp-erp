# Projeto Korp

Este repositório contém **duas versões do mesmo projeto** (serviço Go +
NGINX + Prometheus + Grafana, tudo automatizado via Ansible), cada uma
demonstrando um jeito diferente de provisionar o ambiente com Ansible.
São o mesmo serviço e a mesma stack — o que muda é **onde e como o
Ansible roda**.

## As duas pastas

### [`all-in-one/`](all-in-one/) — control node = managed node

O Ansible roda **dentro da própria máquina** que vai executar a stack
(`ansible_connection=local`). É o cenário mais simples: uma única VM que
instala o Docker nela mesma, sobe os containers nela mesma, e valida nela
mesma. Não depende de rede entre máquinas nem de chave SSH externa.

**Motivo:** demonstração rápida, ambiente único (ex: uma VM Ubuntu
ou WSL2 isolado), ou quando você só tem uma máquina disponível.

**Comando único** (rodado de dentro da própria VM alvo, depois de dar
`git clone` nela):
```bash
cd all-in-one/ansible
ansible-playbook -i inventory.ini playbook.yml -K
```

Guia completo (linguagem simples): [`all-in-one/README.md`](all-in-one/README.md)

### [`by-target/`](by-target/) — control node ≠ managed node

O Ansible roda numa VM (**control node**) e se conecta via **SSH** para
provisionar uma VM diferente (**managed node**, definida em
`ansible/inventory.ini`). É o modelo real de como o Ansible é usado em
produção: um único ponto de controle gerenciando servidores remotos. Os
arquivos do projeto (`app/`, `nginx/`, `monitoring/`, `docker-compose.yml`)
vivem no control node e são sincronizados para o managed node
automaticamente pelo playbook, antes do build.

**Motivo:** simular/demonstrar um cenário produtivo de verdade, com
separação entre quem orquestra (control node) e quem executa (managed
node) — inclusive com distribuições Linux diferentes entre as duas
pontas (o control node deste repo é Ubuntu; o managed node configurado é
Amazon Linux).

**Comando único** (rodado do **control node**, depois de dar `git clone`
nele — o managed node não precisa de nada instalado manualmente, nem
Ansible):
```bash
cd by-target/ansible
ansible-playbook -i inventory.ini playbook.yml
```

Guia completo (linguagem simples): [`by-target/README.md`](by-target/README.md)

## Resumo rápido

| | `all-in-one/` | `by-target/` |
|---|---|---|
| Onde o Ansible roda | Na própria VM alvo | Numa VM separada (control node) |
| Como alcança o alvo | `connection=local` | SSH (`inventory.ini`) |
| Arquivos do projeto | Já estão na VM alvo | Sincronizados do control node pro managed node pelo playbook |
| Precisa de chave SSH extra | Não | Sim, a do managed node, presente no control node |
| Mais parecido com | Um ambiente de teste/demo isolado | Uso real de Ansible em produção |

Em ambos os casos, o requisito "ambiente totalmente provisionado com um
único comando Ansible" é atendido — a diferença é só a topologia de onde
esse comando roda.
