<!-- Versão em português. Mantenha este arquivo sincronizado com o README.md (versão principal). -->

[English](README.md) | **Português**

# My Software Factory Manager

Gerenciador de clientes e projetos: uma plataforma centralizada para gestão de clientes, projetos e demandas de desenvolvimento de software.

> **Status:** em fase de definição. Ainda não há código; as seções marcadas como _A definir_ serão preenchidas conforme o projeto avança.

## Sumário

- [Sobre o projeto](#sobre-o-projeto)
- [Arquitetura](#arquitetura)
- [Tecnologias](#tecnologias)
- [Build e deploy](#build-e-deploy)
- [Backup e manutenção](#backup-e-manutenção)
- [Documentação](#documentação)
- [Licença](#licença)

## Sobre o projeto

### Objetivo

O objetivo é criar uma plataforma centralizada para gestão de clientes, projetos e demandas de desenvolvimento de software, funcionando como o centro operacional de uma consultoria.

A ideia é ter um ambiente para administrar todo o ciclo de vida de um cliente e seus projetos, desde o primeiro contato e o levantamento de necessidades até a entrega, a validação e a manutenção das soluções.

Mais do que um simples gerenciador de tarefas, o sistema deve funcionar como uma espécie de escritório digital inteligente, organizando as informações, os documentos, a comunicação, as demandas e as atividades de cada projeto.

### Problema que resolve

_A definir._

<!-- Descreva a dor atual: onde as informações de clientes e projetos ficam hoje, o que se perde ou se duplica, e quem sofre com isso. -->

## Arquitetura

### Visão geral

Visão preliminar, a confirmar:

```mermaid
flowchart LR
    user([Usuário]) --> web[Frontend<br/>React]
    web --> api[Backend / API<br/>Rust]
    api --> db[(PostgreSQL)]
    api --> storage[(MinIO<br/>armazenamento S3)]
```

<!-- Descreva a responsabilidade de cada componente e como eles se comunicam. Atualize o diagrama quando a arquitetura for confirmada. -->

### Decisões de arquitetura

As decisões relevantes e difíceis de reverter são registradas como ADRs (Architecture Decision Records) em `docs/adr/`, um arquivo por decisão. Este README traz apenas a visão geral e o índice abaixo.

_Nenhuma decisão registrada ainda._

<!-- Uma linha por ADR, por exemplo:
- [0001 - Título](docs/adr/0001-titulo.md): resumo em uma linha
-->

## Tecnologias

| Tecnologia | Papel no projeto |
| --- | --- |
| [Rust](https://www.rust-lang.org/) | Backend e API |
| [React](https://react.dev/) | Frontend web |
| [PostgreSQL](https://www.postgresql.org/) | Banco de dados relacional |
| [MinIO](https://min.io/) | Armazenamento de objetos compatível com S3 (documentos e arquivos) |
| [Docker](https://www.docker.com/) | Empacotamento e execução dos serviços em containers |

<!-- Inclua novas tecnologias nesta tabela à medida que forem adotadas. -->

## Build e deploy

### Pré-requisitos

| Ferramenta | Versão mínima |
| --- | --- |
| Docker | _A definir_ |
| Rust | _A definir_ |
| Node.js | _A definir_ |

### Configuração

_A definir._

<!-- Liste as variáveis de ambiente (conexão com o PostgreSQL, credenciais e buckets do MinIO etc.) e mantenha um `.env.example` versionado, sem segredos reais. -->

### Ambiente de desenvolvimento

_A definir._

<!-- Comandos para subir o ambiente local do zero: clonar, configurar, iniciar os serviços e acessar a aplicação. -->

### Build

_A definir._

<!-- Como gerar os artefatos de produção: binário do backend, bundle do frontend, imagens Docker. -->

### Deploy

_A definir._

<!-- Onde a aplicação roda, como publicar uma nova versão e como reverter para a anterior. -->

## Backup e manutenção

### Backup

| O quê | Como | Frequência | Retenção |
| --- | --- | --- | --- |
| Banco de dados (PostgreSQL) | _A definir_ | _A definir_ | _A definir_ |
| Arquivos (MinIO) | _A definir_ | _A definir_ | _A definir_ |

<!-- Informe também onde os backups ficam armazenados, de preferência fora do servidor da aplicação. -->

### Restauração

_A definir._

<!-- Passo a passo para restaurar o banco e os arquivos a partir de um backup. Teste a restauração periodicamente: backup que nunca foi restaurado não é garantia. -->

### Manutenção

_A definir._

<!-- Rotinas recorrentes: migrações de banco, atualização de dependências e imagens Docker, rotação de credenciais, monitoramento e logs. -->

## Documentação

- `docs/adr/`: decisões de arquitetura (ADRs).
- `GLOSSARY.md`: glossário dos termos do domínio.
- [`docs/agents/`](docs/agents/): configuração usada pelos agentes de IA neste repositório.

`docs/adr/` e `GLOSSARY.md` são criados conforme o domínio e a arquitetura forem definidos.

## Licença

_A definir._
