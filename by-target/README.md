# Como Rodar o Projeto Korp (control node → managed node)

Este guia explica, em linguagem simples, como colocar o projeto no ar. Não é
preciso entender de programação para seguir os passos — seguindo no modelo copiar e colar,
a seguir os comandos indicados.

Essa versão do projeto simula um cenário real de infraestrutura: você roda
o comando em **uma máquina** (o "control node"), e ela mesma se conecta,
via SSH, para instalar e ligar tudo em **outra máquina** (o "managed
node" — a VM que vai realmente rodar o serviço). Simulando um ambiente de produção
que já usa Ansible, gerenciando servidores remotos a partir de um
único ponto.

## O que é esse projeto

É um pequeno site/serviço que, quando acessado, responde com um "cartão de
visitas" digital (nome do projeto e o horário atual). Junto dele sobem
também duas ferramentas de monitoramento (Prometheus e Grafana), que
mostram gráficos de quantas vezes o serviço foi acessado e se ele está
"no ar".

Tudo isso é organizado e ligado automaticamente por uma ferramenta chamada
**Ansible**, com um único comando — sem precisar configurar nada manualmente.

## O que você precisa ter antes de começar

- **Duas máquinas virtuais (VMs) já ligadas:**
  - o **control node** — de onde você vai rodar o comando (este guia
    assume Ubuntu Linux);
  - o **managed node** — a VM alvo, onde o serviço de fato vai subir.
    Nesta versão do projeto, ela já está definida no arquivo
    `ansible/inventory.ini` (endereço ex: `172.17.47.192`, usuário
    `ec2-user`).

- **As duas VMs precisam se enxergar na rede** (o control node precisa
  conseguir acessar o managed node por SSH). Se você já testou
  `ssh ec2-user@172.17.47.192` a partir do control node e funcionou,
  está tudo certo.

- **A chave SSH do managed node precisa estar dentro do control node**,
  no caminho `~/.ssh/al2023` (é o caminho já configurado no
  `inventory.ini`). Se você ainda não colocou a chave lá, copie-a antes
  de continuar:
  ```bash
  # a partir do seu computador, envie a chave para o control node:
  scp -i /caminho/da/sua/chave-do-control-node ~/.ssh/al2023 usuario@endereco-do-control-node:~/.ssh/al2023
  ```
  E, já dentro do control node, ajuste a permissão (o SSH recusa chaves
  com permissão aberta demais):
  ```bash
  chmod 600 ~/.ssh/al2023
  ```

- **Três programinhas precisam já estar instalados dentro do control
  node antes de começar**, porque são eles que "entendem" os comandos dos
  passos seguintes:
  - **Git** — para baixar o projeto;
  - **Python** — o Ansible é escrito nessa linguagem e não funciona sem ela;
  - **Ansible** — a ferramenta que lê as instruções do projeto e monta
    tudo sozinha, remotamente, no managed node.

  O Passo 1 abaixo instala os três de uma vez, então não se preocupe se
  ainda não tiver nenhum deles.

O único programa que **não** precisa instalar antes é o **Docker** — esse
é instalado automaticamente pelo próprio processo, direto no managed node.

## Passo 1 — Acessar o control node e preparar as ferramentas

Abra um terminal e entre no control node via SSH (troque pelo usuário e
endereço da sua VM):

```bash
ssh usuario@endereco-do-control-node
```

Já dentro do control node, instale o Git, o Python e o Ansible (copie e
cole o bloco inteiro de uma vez):

```bash
sudo apt update
sudo apt install -y git python3 ansible
```

Isso só precisa ser feito uma vez. Se a VM já tiver algum desses
programas, o comando acima simplesmente confirma que já está tudo certo,
sem dar erro.

## Passo 2 — Baixar o projeto

Ainda no control node, copie o projeto para lá:

```bash
git clone https://github.com/joaovitorsbz/korp-erp.git
cd korp-erp/by-target/ansible
```

## Passo 3 — Rodar o comando único que monta tudo

Este é o comando principal. Rodado do control node, ele se conecta no
managed node via SSH, instala tudo que falta lá, organiza os serviços e
deixa o projeto pronto para uso:

```bash
ansible-playbook -i inventory.ini playbook.yml
```

Na maioria das VMs (como a EC2 com usuário `ec2-user` usada aqui), esse
comando já funciona sem pedir senha, porque esse tipo de usuário já vem
configurado com permissão de administrador liberada. Se em algum momento
ele reclamar de permissão, rode com `-K` no final:

```bash
ansible-playbook -i inventory.ini playbook.yml -K
```

Aí sim vai aparecer a mensagem `BECOME password:` — digite a senha de
administrador (sudo) **do managed node** (não do control node) e aperte
Enter (a senha não aparece na tela enquanto você digita — é normal).

O processo pode levar alguns minutos dependendo dos recursos disponíveis
no managed node. Ele vai, na ordem, tudo dentro do managed node:

1. Instalar o Docker (o programa que "empacota" e roda os serviços);
2. Criar uma rede interna para os serviços conversarem entre si;
3. Copiar os arquivos do projeto do control node para o managed node;
4. Montar e ligar o serviço principal, o site de proxy (NGINX) e as
   ferramentas de monitoramento (Prometheus e Grafana);
5. Testar sozinho se tudo subiu certo, e mostrar o resultado desse teste
   no final da tela.

Se no final aparecer uma mensagem com `"nome": "Projeto Korp"` e um
horário, deu tudo certo.

## Passo 4 — Conferir no navegador

Com tudo no ar, abra um navegador (Chrome, Firefox, etc.) no seu
computador e acesse o **endereço do managed node** (não é mais
`localhost`, já que o serviço está rodando em outra máquina):

- **`http://172.17.47.192/`** — página inicial do projeto, com um botão
  para testar o serviço.
- **`http://172.17.47.192:3000`** — o painel de monitoramento (Grafana).
  Login: `admin` / Senha: `admin`. O gráfico com as métricas do projeto já
  vem pronto, sem precisar configurar nada.

> Se o managed node estiver numa rede que só o control node alcança
> (não o seu computador), acesse esses endereços a partir do próprio
> control node, ou de um navegador que tenha rota até `172.17.47.192`.

## Se algo der errado

- **"Command not found: ansible-playbook" (ou "git")** — algum dos três
  programas do Passo 1 não ficou instalado **no control node**. Repita o
  comando do Passo 1:
  ```bash
  sudo apt update
  sudo apt install -y git python3 ansible
  ```
  Depois repita o Passo 3.

- **"Permission denied (publickey)" ou erro de SSH** — a chave
  `~/.ssh/al2023` não está no control node, ou está com a permissão
  errada. Revise a seção de pré-requisitos acima (`scp` + `chmod 600`).

- **Pediu a senha (`BECOME password`) e não aceitou** — é a senha de
  administrador do **managed node** (a VM `172.17.47.192`), não a do
  control node nem de outro sistema.

- **A porta 80 já está em uso no managed node** — algum outro programa
  lá já está usando essa porta. Rode o comando do Passo 3 assim, usando
  outra porta:
  ```bash
  ansible-playbook -i inventory.ini playbook.yml --extra-vars "korp_http_port=8081"
  ```
  E depois acesse `http://172.17.47.192:8081/` no lugar da porta 80.
