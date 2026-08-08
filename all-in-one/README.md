# Como Rodar o Projeto Korp

Este guia explica, em linguagem simples, como colocar o projeto no ar. Não é
preciso entender de programação para seguir os passos — só copiar e colar
os comandos indicados.

## O que é esse projeto

É um pequeno site/serviço que, quando acessado, responde com um "cartão de
visitas" digital (nome do projeto e o horário atual). Junto dele sobem
também duas ferramentas de monitoramento (Prometheus e Grafana), que
mostram gráficos de quantas vezes o serviço foi acessado e se ele está
"no ar".

Tudo isso é organizado e ligado automaticamente por uma ferramenta chamada
**Ansible**, com um único comando — sem precisar configurar nada manualmente.

## O que você precisa ter antes de começar

- **Uma máquina virtual (VM) com Ubuntu Linux instalado e ligada.**
  Pode ser criada em qualquer programa de virtualização (VirtualBox,
  VMware, Hyper-V) ou em um servidor na nuvem. O importante é que:
  - o sistema seja **Ubuntu** (versão 22.04 ou mais recente);
  - você consiga **acessá-la por SSH** (o "controle remoto" por linha de
    comando que se usa para acessar servidores Linux);
  - você tenha o **usuário e a senha** (ou a chave de acesso) dessa VM em
    mãos, pois vai ser pedida durante o processo.

- **Três programinhas precisam já estar instalados dentro dessa VM antes
  de começar**, porque são eles que "entendem" os comandos dos passos
  seguintes:
  - **Git** — para baixar o projeto;
  - **Python** — o Ansible é escrito nessa linguagem e não funciona sem ela;
  - **Ansible** — a ferramenta que lê as instruções do projeto e monta
    tudo sozinha.

  O Passo 1 abaixo instala os três de uma vez, então não se preocupe se
  ainda não tiver nenhum deles.

O único programa que **não** precisa instalar antes é o **Docker** — esse
é instalado automaticamente pelo próprio processo, junto com o resto.

## Passo 1 — Acessar a VM e preparar as ferramentas

Abra um terminal e entre na sua máquina virtual via SSH (troque pelo
usuário e endereço da sua VM):

```bash
ssh usuario@endereco-da-sua-vm
```

Já dentro da VM, instale o Git, o Python e o Ansible (copie e cole o bloco
inteiro de uma vez):

```bash
sudo apt update
sudo apt install -y git python3 ansible
```

Isso só precisa ser feito uma vez. Se a VM já tiver algum desses
programas, o comando acima simplesmente confirma que já está tudo certo,
sem dar erro.

## Passo 2 — Baixar o projeto

Ainda na VM, copie o projeto para lá com o comando abaixo (troque a URL
pelo endereço do repositório no GitHub):

```bash
git clone https://github.com/SEU-USUARIO/korp.git
cd korp/ansible
```

## Passo 3 — Rodar o comando único que monta tudo

Este é o comando principal. Ele instala tudo que falta, organiza os
serviços e deixa o projeto pronto para uso:

```bash
ansible-playbook -i inventory.ini playbook.yml -K
```

Assim que você apertar Enter, vai aparecer a mensagem `BECOME password:`.
Digite a **senha de administrador (sudo) da própria VM** e aperte Enter
(a senha não aparece na tela enquanto você digita — é normal, continue e
aperte Enter no final).

O processo leva alguns minutos. Ele vai, na ordem:

1. Instalar o Docker (o programa que "empacota" e roda os serviços);
2. Criar uma rede interna para os serviços conversarem entre si;
3. Montar e ligar o serviço principal, o site de proxy (NGINX) e as
   ferramentas de monitoramento (Prometheus e Grafana);
4. Testar sozinho se tudo subiu certo, e mostrar o resultado desse teste
   no final da tela.

Se no final aparecer uma mensagem com `"nome": "Projeto Korp"` e um
horário, deu tudo certo.

## Passo 4 — Conferir no navegador

Com tudo no ar, abra um navegador (Chrome, Firefox, etc.) e acesse os
seguintes endereços — troque `localhost` pelo endereço da sua VM, se
estiver acessando de outro computador:

- **`http://localhost/`** — página inicial do projeto, com um botão para
  testar o serviço.
- **`http://localhost:3000`** — o painel de monitoramento (Grafana).
  Login: `admin` / Senha: `admin`. O gráfico com as métricas do projeto já
  vem pronto, sem precisar configurar nada.

## Se algo der errado

- **"Command not found: ansible-playbook" (ou "git")** — algum dos três
  programas do Passo 1 não ficou instalado. Repita o comando do Passo 1:
  ```bash
  sudo apt update
  sudo apt install -y git python3 ansible
  ```
  Depois repita o Passo 3.

- **Pediu a senha e não aceitou** — confirme que é a senha do seu usuário
  na VM (a mesma que você usaria com o comando `sudo`), não a senha de
  outro sistema.

- **A porta 80 já está em uso** — algum outro programa na VM já está
  usando essa porta. Rode o comando do Passo 3 assim, usando outra porta:
  ```bash
  ansible-playbook -i inventory.ini playbook.yml -K --extra-vars "korp_http_port=8081"
  ```
  E depois acesse `http://localhost:8081/` no lugar de `http://localhost/`.
