<!-- Versão em português. Mantenha este arquivo sincronizado com o README.md (versão principal). -->

[English](README.md) | **Português**

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/design/brand/mark-on-dark.svg">
  <img src="docs/design/brand/mark.svg" alt="" width="72" height="72">
</picture>

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

### Quem usa e como

```mermaid
flowchart TB
    subgraph callers ["Quem usa"]
        person(["Pessoa<br/>um usuário cadastrado"])
        coding(["Agentes de código<br/>Claude Code, Codex, OpenCode, pi"])
        custom(["Agentes próprios<br/>serviços separados, criados depois"])
    end

    subgraph interfaces ["Interfaces"]
        web["App web<br/>SPA React no navegador"]
        mcp["Servidor MCP<br/>um conjunto escolhido de ferramentas"]
        rest["API REST<br/>todas as operações"]
    end

    subgraph backend ["Backend"]
        usecases["Casos de uso<br/>conferem quem está agindo<br/>e o que é permitido"]
        worker["Worker<br/>trabalhos em segundo plano"]
    end

    subgraph stores ["Onde os dados ficam"]
        files[("Armazenamento<br/>de arquivos")]
        db[("PostgreSQL<br/>dados, sessões, trabalhos,<br/>registro de auditoria")]
    end

    subgraph outside ["Serviços externos"]
        idp(["Google, Microsoft, GitHub"])
        llm(["Provedores de LLM"])
        mail(["E-mail"])
    end

    person -->|"entra por um provedor<br/>ou código de login"| web
    web -->|"cookie de sessão"| rest
    coding -->|"token de acesso<br/>com escopos"| mcp
    custom -->|"token de acesso<br/>com escopos"| rest
    rest --> usecases
    mcp --> usecases
    web <-->|"arquivos, por<br/>URLs assinadas"| files
    usecases -->|"lê, grava,<br/>salva trabalhos"| db
    usecases -->|"confere o login"| idp
    usecases --> llm
    db -->|"trabalhos"| worker
    worker --> llm
    worker -->|"código de login<br/>e outros e-mails"| mail
```

Três tipos de consumidor chegam ao sistema, cada um pela sua interface:

- **Uma pessoa** usa o app web no navegador. Ela entra pelo Google, pela Microsoft ou pelo GitHub, ou com um código de login enviado por e-mail, e só entra quem é usuário cadastrado.
- **Agentes de código** usam o servidor MCP, que oferece um conjunto escolhido de ferramentas: atualizar o status de um projeto, consultar clientes e horas, buscar uma credencial.
- **Agentes próprios**, criados depois como serviços separados, usam a API REST.

Um agente entra com um token de acesso que um usuário criou para ele, limitado por escopos, e o que ele faz fica registrado como ação desse usuário. Todas as interfaces terminam nos mesmos casos de uso, que decidem se a ação é permitida, então as regras são as mesmas para qualquer consumidor.

Os dados são lidos e gravados no PostgreSQL. Os arquivos vão direto do navegador para o armazenamento e de volta, por URLs assinadas emitidas pela API. O que ninguém espera na tela é salvo como um trabalho e executado pelo worker: envio de e-mail, processamento de arquivos e chamadas a provedores de LLM.

Os tipos de usuário, e o que cada um pode fazer, ainda não foram modelados.

## Arquitetura

### Visão geral

Decidido até aqui (veja os ADRs abaixo):

```mermaid
flowchart TB
    actions(["GitHub Actions<br/>constrói e publica<br/>as imagens"])
    agents(["Agentes de código"])
    spa["Navegador: SPA React<br/>TanStack Router, DB,<br/>Query e Form<br/>Zustand, react-i18next<br/>shadcn/ui, Tailwind CSS<br/>cliente gerado pelo orval"]

    subgraph server ["Servidor: Docker Compose"]
        caddy["Caddy<br/>HTTPS, arquivos do SPA,<br/>proxy reverso"]
        subgraph binary ["Um único binário Go"]
            api["API<br/>Huma (REST, OpenAPI)<br/>MCP Go SDK<br/>SCS, pgx, sqlc<br/>AWS SDK para Go v2"]
            worker["Worker<br/>River"]
        end
        migrate["Migrations<br/>goose"]
        db[("PostgreSQL 18<br/>dados, sessões,<br/>fila de trabalhos")]
        telemetry["Telemetria<br/>por enquanto só em desenvolvimento<br/>OpenTelemetry Collector<br/>Grafana otel-lgtm"]
    end

    subgraph outside ["Serviços externos"]
        idp(["Login<br/>Google, Microsoft, GitHub<br/>Dex em desenvolvimento"])
        llm(["Provedores de LLM<br/>formato Chat Completions"])
        mail(["E-mail<br/>Resend em produção<br/>Mailpit nos demais ambientes"])
        storage[("Armazenamento S3<br/>Backblaze B2 em produção<br/>Garage nos demais ambientes")]
    end

    actions -->|"SSH"| server
    agents -->|"MCP"| caddy
    spa -->|"HTTPS"| caddy
    caddy -->|"REST e MCP"| api
    binary --> db
    migrate --> db
    binary & caddy -->|"OTLP"| telemetry
    api -->|"login"| idp
    binary --> llm
    worker --> mail
    binary -->|"arquivos"| storage
    db -->|"backups,<br/>pgBackRest"| storage
    spa -->|"arquivos, por<br/>URLs assinadas"| storage
    db ~~~ idp
    db ~~~ llm
    db ~~~ mail
```

O worker executa os trabalhos em segundo plano que a API grava no PostgreSQL: envio de e-mail, processamento de arquivos e chamadas a provedores de LLM que ninguém espera na tela. O navegador envia e baixa arquivos direto do armazenamento, com URLs assinadas emitidas pela API.

Em produção os arquivos ficam no Backblaze B2. Em desenvolvimento e na homologação ficam no Garage, atrás de um endereço do Caddy.

A telemetria passa pelo coletor: traces e métricas da API, do worker e do Caddy, a saída de todos os containers e as métricas do servidor, do PostgreSQL e do armazenamento. Ele repassa tudo ao Grafana otel-lgtm, que guarda por 30 dias e envia os alertas por e-mail. O Grafana tem um endereço próprio atrás do Caddy, com login pelo GitHub. Por enquanto os dois containers de telemetria rodam só em desenvolvimento: a produção passa a tê-los quando alguém além do mantenedor usar o sistema, e até lá um monitor hospedado fora do servidor a vigia.

O sistema roda em três ambientes: desenvolvimento na máquina do mantenedor, homologação (staging) em um servidor criado para um pull request e destruído depois, e produção em um servidor sempre ligado.

Frontend e backend ficam neste único repositório. As pastas ainda sem uso têm apenas um arquivo `.gitkeep`, e os arquivos de configuração da raiz entram junto com o primeiro código:

```
.
├── apps/
│   ├── api/          # backend e API em Go
│   └── web/          # o app React, com o design system em src/ui
├── deploy/           # Caddyfile, arquivos Compose dos servidores, cloud-init
├── .github/          # workflows da pipeline
├── docs/
│   ├── adr/          # decisões de arquitetura
│   ├── agents/       # configuração para agentes de IA
│   └── design/       # análise das referências visuais
├── .scratch/         # issues locais
├── compose.yaml      # desenvolvimento
└── Taskfile.yml
```

`apps/web` é o único pacote JavaScript, executado pelo Vite+. Ele tem o próprio `package.json`, catálogo e lockfile, e a raiz não tem nenhum. Um Taskfile na raiz coordena as tarefas de Go e de frontend: desenvolvimento, build, testes e geração de código. Não é usado orquestrador de monorepo à parte.

As telas são construídas direto no app, sobre o design system, sem protótipo antes. O design system está escrito em [`DESIGN.md`](DESIGN.md), que é a referência para todas as telas, e os componentes ficam em `apps/web/src/ui`.

O sistema é multilíngue: inglês é a língua padrão, com suporte também a português do Brasil e espanhol, e cada usuário escolhe a sua. Adicionar uma língua é uma mudança pequena de código, sem alterar a estrutura do banco de dados.

<!-- Descreva a responsabilidade de cada componente e como eles se comunicam. Atualize o diagrama quando a arquitetura for confirmada. -->

### Decisões de arquitetura

As decisões relevantes e difíceis de reverter são registradas como ADRs (Architecture Decision Records) em `docs/adr/`, um arquivo por decisão. Este README traz apenas a visão geral e o índice abaixo.

- [0001 - Go no backend](docs/adr/0001-go-for-the-backend.md): Go em vez de TypeScript e Rust, porque agentes escrevem o código e o ciclo de build precisa ser curto
- [0002 - React como SPA, sem framework full-stack](docs/adr/0002-react-spa-without-a-full-stack-framework.md): Vite+ e TanStack Router atrás do Caddy, sem Next.js nem TanStack Start
- [0003 - Contrato OpenAPI gerado do Go](docs/adr/0003-openapi-contract-generated-from-go.md): os tipos Go são a fonte da verdade da API REST, via Huma
- [0004 - Provedores de LLM e agentes via HTTP](docs/adr/0004-llm-providers-and-agents-over-http.md): uma interface interna para os provedores; agentes são serviços separados que agem apenas pela API
- [0005 - Multilíngue desde o início](docs/adr/0005-multilingual-from-the-start.md): inglês, português do Brasil e espanhol; o front traduz os códigos de erro da API; conteúdo de configuração tem uma versão por língua, texto livre é traduzido sob demanda
- [0006 - Design system validado em um protótipo descartável](docs/adr/0006-design-system-proven-in-a-disposable-prototype.md): shadcn/ui em `packages/ui`, compartilhado pelo produto e por um protótipo clicável; substituído pelo ADR-0025, que mantém o shadcn/ui e o estilo do Zed
- [0007 - Backend organizado por área de domínio](docs/adr/0007-backend-organized-by-domain-area.md): uma árvore de pacotes por área de domínio com dependências apontando para dentro, um caso de uso para cada operação, DDD tático seletivo e parte do object calisthenics
- [0008 - Casos de uso atendidos por vários adaptadores de entrada](docs/adr/0008-use-cases-served-by-several-inbound-adapters.md): uma API REST completa e um servidor MCP com um conjunto escolhido de ferramentas, ambos no backend Go; tokens de acesso com escopos; GraphQL, gRPC e WebSocket adiados
- [0009 - Convenções da API REST](docs/adr/0009-rest-api-conventions.md): snake_case, operações nomeadas sem PATCH, entrada estrita, erros em problem+json, paginação por cursor, sem versionamento enquanto todo consumidor sair junto com o backend
- [0010 - Identificadores sequenciais expostos via Sqids](docs/adr/0010-sequential-identifiers-exposed-through-sqids.md): inteiros por dentro, prefixo de tipo mais codificação Sqids por fora (`prj_Xk9fQ2mB`), com alfabetos fixos no código
- [0011 - Login por provedores externos ou por código de login](docs/adr/0011-login-through-external-providers-with-a-server-side-session.md): Google, Microsoft e GitHub, ou um código de seis dígitos enviado por e-mail; sessão no servidor atrás de um cookie opaco, apenas usuários cadastrados
- [0012 - OpenTelemetry e um registro de auditoria separado](docs/adr/0012-opentelemetry-for-observability-and-a-separate-audit-log.md): traces e métricas enviados por um coletor a um container Grafana, logs lidos da saída dos containers, alertas por e-mail; operações sensíveis gravadas em um registro de auditoria somente de acréscimo
- [0013 - Dados do frontend em coleções e consultas por cursor](docs/adr/0013-frontend-data-in-collections-and-cursor-queries.md): TanStack DB para conjuntos limitados, TanStack Query com o cursor da API para o resto, Zustand só para estado de interface, cliente e schemas Zod gerados pelo orval, TanStack Form
- [0014 - Frontend organizado por área de domínio](docs/adr/0014-frontend-organized-by-domain-area-and-tested-through-screens.md): uma pasta por área de domínio com um módulo de dados que é o único a chamar a API; os testes renderizam telas inteiras em um navegador real com a API simulada
- [0015 - SQL escrito à mão e verificado pelo sqlc, e migrations só para frente](docs/adr/0015-hand-written-sql-checked-by-sqlc-and-forward-only-migrations.md): pgx e sqlc, migrations do goose sem caminho de volta aplicadas em um passo próprio, três usuários do banco, PostgreSQL como único banco de dados, sem cache e sem pool de conexões externo
- [0016 - Trabalhos em segundo plano no PostgreSQL, executados por um comando worker](docs/adr/0016-background-jobs-in-postgresql-run-by-a-worker-command.md): River, o trabalho gravado na transação do seu caso de uso, um comando `worker` do mesmo binário, poucas tentativas e depois uma falha visível
- [0017 - Arquivos no Garage, acessados direto pelo navegador](docs/adr/0017-files-in-garage-reached-directly-by-the-browser.md): Garage no lugar do MinIO, apenas a API S3 pelo AWS SDK para Go v2, uploads e downloads com URLs assinadas de curta duração
- [0018 - Valores de credenciais cifrados pela API](docs/adr/0018-credential-values-encrypted-by-the-api.md): AES-256-GCM no binário Go com uma chave vinda da configuração do deploy, nunca no banco; uma credencial em arquivo é guardada do mesmo modo
- [0019 - Três ambientes, com a homologação em um servidor criado sob demanda](docs/adr/0019-three-environments-with-staging-on-a-server-created-on-demand.md): desenvolvimento na máquina, um servidor de homologação criado para um pull request rotulado e populado pelo seed a cada publicação, produção em um servidor da Hetzner; o GitHub Actions publica por SSH, com os segredos nos Environments do GitHub
- [0020 - PostgreSQL com backup pelo pgBackRest, e arquivos de produção no Backblaze B2](docs/adr/0020-postgresql-backed-up-by-pgbackrest-and-production-files-in-backblaze-b2.md): backups contínuos e cifrados em outro provedor, com um teste mensal de restauração, e os arquivos de produção em um bucket com versionamento; o Garage continua em desenvolvimento e na homologação
- [0021 - Um repositório público sob a AGPLv3](docs/adr/0021-a-public-repository-under-agplv3.md): público no plano gratuito do GitHub, com as regras que isso impõe aos commits e à pipeline
- [0022 - Um registro com ciclo de vida é uma máquina de estados no domínio](docs/adr/0022-a-record-with-a-lifecycle-is-a-state-machine-in-the-domain.md): um conjunto fixo de status e uma tabela de transições por agregado, em Go puro; as ações permitidas saem da mesma tabela, e só o nome de um status é conteúdo de configuração
- [0023 - Uma release dá nome a um marco que já está em produção](docs/adr/0023-a-release-names-a-milestone-already-in-production.md): uma versão para o sistema inteiro em um arquivo `VERSION`; uma release é um pull request que muda esse arquivo e o `CHANGELOG.md`, e a pipeline cria a tag depois da publicação
- [0024 - O primeiro administrador é criado em uma configuração inicial aberta por um código de configuração](docs/adr/0024-the-first-administrator-is-created-in-a-setup-opened-by-a-setup-code.md): uma instalação nova não tem usuário; um código de configuração emitido por um comando no servidor abre a configuração inicial, que registra a consultoria e o primeiro administrador de uma vez
- [0025 - O design system fica no produto, e não há protótipo](docs/adr/0025-the-design-system-lives-in-the-product-and-there-is-no-prototype.md): o protótipo, o pacote compartilhado e a galeria são removidos; uma tela é construída direto em `apps/web`, que guarda os componentes e é o único pacote JavaScript
- [0026 - Um papel é um registro que concede escopos fixos no código](docs/adr/0026-a-role-is-a-record-that-grants-scopes-fixed-in-the-code.md): a consultoria cria seus próprios papéis a partir de uma lista de escopos fixa no código; um usuário pode ter vários, uma sessão age com os escopos dos papéis do seu usuário, e o papel de administrador sempre concede todos os escopos

## Tecnologias

| Tecnologia | Papel no projeto |
| --- | --- |
| [Go](https://go.dev/) | Backend e API |
| [Huma](https://huma.rocks/) | Framework HTTP que gera a especificação OpenAPI a partir dos tipos Go |
| [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) | Servidor MCP dentro do backend, pelo qual agentes de código operam o sistema |
| [Sqids](https://sqids.org/go) | Identificadores públicos curtos, codificados a partir dos sequenciais internos |
| [SCS](https://github.com/alexedwards/scs) | Sessões no servidor, guardadas no PostgreSQL |
| [env](https://github.com/caarlos0/env) | Lê a configuração das variáveis de ambiente para uma única struct tipada |
| [pgx](https://github.com/jackc/pgx) | Driver do PostgreSQL e o pool de conexões dentro da API |
| [sqlc](https://sqlc.dev/) | Gera código Go tipado a partir de SQL escrito à mão |
| [goose](https://github.com/pressly/goose) | Aplica as migrations do banco, só para frente |
| [pgtestdb](https://github.com/peterldowns/pgtestdb) | Dá a cada teste de banco o seu próprio banco temporário |
| [River](https://riverqueue.com/) | Trabalhos em segundo plano guardados no PostgreSQL e executados pelo comando `worker`, com um painel web |
| [AWS SDK para Go v2](https://github.com/aws/aws-sdk-go-v2) | Cliente do armazenamento S3; assina as URLs de upload e download |
| [go-mail](https://github.com/wneessen/go-mail) | Envia e-mail por SMTP em desenvolvimento e nos testes |
| [Resend](https://resend.com/) | Envia os e-mails em produção e avisa das devoluções |
| [Mailpit](https://mailpit.axllent.org/) | Container que captura os e-mails enviados em desenvolvimento e nos testes |
| [OpenTelemetry](https://opentelemetry.io/) | Traces e métricas do backend |
| [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) | Container que reúne traces, métricas e logs de todos os serviços e do servidor |
| [Grafana otel-lgtm](https://grafana.com/docs/opentelemetry/docker-lgtm/) | Um container que guarda e mostra a telemetria e envia os alertas |
| [vacuum](https://quobix.com/vacuum/) e [oasdiff](https://github.com/oasdiff/oasdiff) | Lint das convenções da API e relatório de mudanças incompatíveis na especificação OpenAPI |
| [React](https://react.dev/) | Frontend web (single-page app) |
| [Vite+](https://viteplus.dev/) | Ferramentas do frontend em uma única CLI (`vp`): servidor de desenvolvimento e build (Vite, Rolldown), testes (Vitest), lint (Oxlint), formatação (Oxfmt), além do runtime Node.js e do gerenciador de pacotes |
| [TanStack Router](https://tanstack.com/router) | Roteamento tipado no cliente; filtros e paginação ficam na URL |
| [TanStack DB](https://tanstack.com/db) e [TanStack Query](https://tanstack.com/query) | Dados da API no navegador: coleções para conjuntos pequenos, consultas por cursor para o resto |
| [TanStack Form](https://tanstack.com/form) e [Zod](https://zod.dev/) | Formulários, validados com schemas gerados da especificação OpenAPI |
| [Zustand](https://github.com/pmndrs/zustand) | Estado de interface compartilhado entre telas (tema, painéis) |
| [orval](https://orval.dev/) | Gera o cliente TypeScript, os schemas Zod e respostas simuladas a partir da especificação OpenAPI |
| [React Compiler](https://react.dev/learn/react-compiler) | Memoização automática dos componentes no build |
| [Playwright](https://playwright.dev/) e [MSW](https://mswjs.io/) | Testes de tela em navegador real com a API simulada (pelo Vitest) e testes de ponta a ponta |
| [Knip](https://knip.dev/) | Encontra arquivos, exports e dependências sem uso |
| [Chrome DevTools MCP](https://github.com/ChromeDevTools/chrome-devtools-mcp) | Permite aos agentes de código investigar a performance em um Chrome real |
| [react-i18next](https://react.i18next.com/) | Traduções da interface (inglês, português do Brasil, espanhol) |
| [shadcn/ui](https://ui.shadcn.com/) | Base do design system: o código dos componentes é copiado para `apps/web/src/ui`, sobre os primitivos do [Base UI](https://base-ui.com/) |
| [Tailwind CSS](https://tailwindcss.com/) | Estilos e variáveis de estilo do design system (temas claro e escuro) |
| [PostgreSQL](https://www.postgresql.org/) | Banco de dados relacional (versão 18), o único do sistema |
| [Garage](https://garagehq.deuxfleurs.fr/) | Armazenamento de objetos compatível com S3 (documentos e arquivos) em desenvolvimento e na homologação |
| [Backblaze B2](https://www.backblaze.com/cloud-storage) | Armazenamento de objetos compatível com S3 em produção: os arquivos e os backups do banco |
| [pgBackRest](https://pgbackrest.org/) | Backup contínuo e cifrado do PostgreSQL, com restauração para um ponto no tempo |
| [Dex](https://dexidp.io/) | Container de identidade por trás do provedor de login local, em desenvolvimento e nos testes |
| [Hetzner Cloud](https://www.hetzner.com/cloud) | O servidor de produção e o servidor de homologação criado sob demanda |
| [GitHub Actions](https://docs.github.com/actions) | A pipeline: verificações em cada pull request, publicação e testes agendados |
| [GitHub CLI](https://cli.github.com/) | Como os agentes acessam o GitHub: pull requests, execuções e segredos |
| [Caddy](https://caddyserver.com/) | HTTPS, arquivos estáticos do frontend, proxy reverso para a API, cabeçalhos de segurança e o limite do código de login e do código de configuração |
| [Docker](https://www.docker.com/) | Empacotamento e execução dos serviços em containers (Docker Compose) |
| [Task](https://taskfile.dev/) | Executor de tarefas do repositório (desenvolvimento, build, testes, geração de código) |
| [pnpm](https://pnpm.io/) | Gerenciador de pacotes do frontend, instalado e executado pelo Vite+ |

<!-- Inclua novas tecnologias nesta tabela à medida que forem adotadas. -->

## Build e deploy

### Pré-requisitos

| Ferramenta | Versão mínima |
| --- | --- |
| Docker | Engine 28.3.2 com Compose 2.38.2 (Docker Desktop 4.43.2) |
| Go | 1.27.4 |
| Vite+ (gerencia Node.js e pnpm) | 1.6.0 |
| Task | 3.54.0 |
| GitHub CLI | 2.104.0 |

Cada mínimo é a versão em uso na máquina do mantenedor, registrada em [`docs/machine-tools.yaml`](docs/machine-tools.yaml). Nenhuma versão anterior foi testada.

O Go instalado só baixa o toolchain que o `go.mod` indica. Todas as outras versões usadas pelo build são fixadas no repositório (veja Versions em [`CODING_STANDARDS.md`](CODING_STANDARDS.md)).

### Configuração

O binário da API lê a configuração apenas de variáveis de ambiente, uma única vez, ao iniciar. Um valor ausente ou inválido encerra o processo antes de ele atender qualquer coisa, com uma linha que cita a variável e nunca o valor. O binário não lê `.env` nem arquivo de configuração: o Task e o Compose carregam as variáveis.

Toda variável está listada no `.env.example` versionado, sem segredos reais.

| Ambiente | Onde ficam os valores |
| --- | --- |
| Desenvolvimento | O `.env.example`, cujos valores funcionam como estão. Um arquivo `.env`, nunca versionado, altera um valor: o Task o lê primeiro |
| Homologação e produção | O Environment do GitHub de cada um: um segredo para o que é secreto, uma variável para o restante. A publicação os grava em um arquivo de variáveis no servidor, que só o usuário de deploy consegue ler |

Uma variável nova precisa ter o valor definido nos dois Environments do GitHub antes do merge. A publicação compara os nomes do `.env.example` com os valores que tem e para, antes de reiniciar qualquer coisa, quando falta algum.

Cada comando do binário exige apenas as variáveis que usa. As que o binário lê hoje:

| Variável | O que é | Padrão |
| --- | --- | --- |
| `API_LISTEN_ADDRESS` | Onde a API escuta, no formato `host:porta`. Um host vazio significa todas as interfaces | Nenhum: a API exige |
| `SHUTDOWN_DRAIN_DEADLINE` | Quanto tempo as requisições em andamento têm para terminar depois que a API recebe a ordem de parar | `20s` |
| `LOG_LEVEL` | O menor nível que o binário registra nos logs: `debug`, `info`, `warn` ou `error` | `info` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | URL base do receptor OTLP/HTTP para onde os traces são exportados, como `http://localhost:4318`. Vazia, nada é exportado | Vazio |
| `DATABASE_URL` | URL de conexão do papel da API, que lê e grava linhas | Nenhum: a API exige |
| `MIGRATION_DATABASE_URL` | URL de conexão do papel de migração, que é dono das tabelas | Nenhum: o `migrate` exige |
| `DATABASE_POOL_MAX_SIZE` | O máximo de conexões que o pool da API abre | `10` |
| `DATABASE_STATEMENT_TIMEOUT` | Por quanto tempo um comando SQL pode rodar | `30s` |
| `DATABASE_IDLE_IN_TRANSACTION_TIMEOUT` | Por quanto tempo uma transação pode ficar ociosa | `1m` |
| `HEALTH_CHECK_TIMEOUT` | Quanto tempo a verificação de uma parte pode levar, incluída a de prontidão | `2s` |

O `.env.example` também guarda o que só o desenvolvimento usa: a porta e as senhas do container do PostgreSQL, e a conexão com que os testes em Go criam seus bancos temporários.

As variáveis do que ainda não foi construído (as credenciais e o bucket do armazenamento, o servidor de e-mail e as demais) estão _A definir_, e cada uma entra com o código que a lê.

### Ambiente de desenvolvimento

Go, Vite+ e Task rodam na máquina, então um teste roda sem passar pelo Docker. `task dev` roda a API, o worker e o Vite. O `compose.yaml` da raiz roda os serviços com que eles conversam:

| Serviço | Em desenvolvimento |
| --- | --- |
| Proxy reverso | Caddy |
| Banco de dados | PostgreSQL |
| Armazenamento de arquivos | Garage |
| E-mail | Mailpit, que captura o que é enviado |
| Login | Dex como provedor local, ou um código de login lido no Mailpit |
| Provedor de LLM | Um substituto escrito neste repositório, com respostas fixas |
| Telemetria | O OpenTelemetry Collector e o Grafana otel-lgtm, em um perfil do Compose desligado por padrão |

O seed preenche o banco com dados fictícios, os mesmos a cada execução. Para trabalhar em um prompt com um provedor de LLM real, altere o `.env` local.

A pipeline chama os mesmos alvos do Task, então uma verificação que falha nela falha do mesmo jeito na máquina.

Da tabela acima, o PostgreSQL é o que roda hoje, ao lado da API e do app. Para iniciá-los a partir de um clone novo, com os pré-requisitos instalados e o Docker em execução, rode na raiz do repositório:

1. `task install`
2. `task services`
3. `task migrate`
4. `task api`
5. `task dev`, em outro terminal

Nada é configurado à mão: os valores do `.env.example` funcionam como estão. A API responde à verificação de atividade em `http://127.0.0.1:8080/livez` e à de prontidão em `http://127.0.0.1:8080/readyz`, que responde 503 até as migrations serem aplicadas.

Os comandos, executados na raiz do repositório:

| Comando | O que faz |
| --- | --- |
| `task install` | Instala as dependências JavaScript. O Vite+ baixa as versões de Node.js e pnpm que o `apps/web/package.json` indica |
| `task services` | Inicia os serviços de desenvolvimento no Docker e espera até ficarem saudáveis. Toda cópia de trabalho do repositório, inclusive um worktree do Git, usa os mesmos containers |
| `task services:stop` | Para os serviços de desenvolvimento e mantém os dados |
| `task migrate` | Aplica as migrations pendentes no banco de desenvolvimento. Uma segunda execução não aplica nada |
| `task api` | Inicia a API em `http://127.0.0.1:8080` |
| `task dev` | Inicia o app em `http://localhost:5173` |
| `task api:build` | Gera o binário da API em `apps/api/bin` |
| `task api:test` | Inicia os serviços e roda os testes em Go. Um teste que precisa do PostgreSQL cria nele o próprio banco temporário. Os argumentos depois de `--` substituem a lista de pacotes |
| `task check` | Roda todas as verificações de um pull request: as de formato, lint e tipos de `apps/web`, as chaves de tradução ausentes e os testes do app e, em seguida, o build da API e os testes em Go |
| `task check:fix` | Roda as verificações de formato, lint e tipos de `apps/web` e corrige o que for possível |

Os comandos do que ainda não foi construído (o worker, o seed, o código gerado) estão _A definir_, e cada um será escrito com o seu código.

<!-- Comandos para subir o ambiente local do zero: clonar, configurar, iniciar os serviços e acessar a aplicação. -->

### Build

- **Backend.** Um binário Go, em uma imagem: a API e os comandos `worker` e `migrate`. As migrations vão embutidas nele.
- **Seed e substituto do provedor de LLM.** Cada um é um pacote `main` próprio no mesmo módulo Go, com a própria imagem. A produção não roda nenhum dos dois.
- **Frontend.** O Vite+ gera os arquivos estáticos de `apps/web`, com uma cópia Brotli ao lado de cada arquivo de texto. O Caddy os serve.
- **Caddy e PostgreSQL.** As duas imagens são construídas neste repositório: a do Caddy com um plugin de limite de requisições, a do PostgreSQL com o pgBackRest.
- **Código gerado.** A pipeline verifica se ele está atualizado: o código Go que o sqlc gera a partir dos arquivos SQL e o cliente que o orval gera a partir da especificação OpenAPI.
- **Versões.** O build instala a partir do `go.sum` e do `pnpm-lock.yaml` e falha quando eles mudariam.
- **Imagens.** O GitHub Actions as constrói, marca cada uma com o hash do commit e envia para o registro de containers do GitHub. A produção só roda imagem construída a partir da `main`. As imagens de uma release levam também o número dela.

O `task api:build` gera o binário da API em `apps/api/bin`. Os outros comandos estão _A definir_, e cada um será escrito com o que ele constrói.

### Deploy

Toda mudança chega à `main` por um pull request, e um merge na `main` é publicado em produção assim que a pipeline passa. O GitHub Actions constrói as imagens, envia para o registro de containers do GitHub e entra no servidor por SSH para baixá-las e reiniciar os serviços do Compose. Os segredos de cada ambiente ficam nos Environments do GitHub, e a publicação os grava em um arquivo de variáveis no servidor (veja o [ADR-0019](docs/adr/0019-three-environments-with-staging-on-a-server-created-on-demand.md)).

Para ver uma branch antes do merge, dê ao pull request dela o rótulo `staging`. Um servidor é criado com a branch, um banco novo e os dados do seed. Ele é destruído quando o rótulo sai ou o pull request fecha, e nunca vive mais de 24 horas.

Voltar a uma versão anterior é um workflow manual que publica um commit anterior da `main`, indicado pelo hash ou por uma release.

Uma release dá um número e notas a um marco que já está em produção. Ela é um pull request que muda o `VERSION` e o `CHANGELOG.md`. Depois que ele é publicado, a pipeline cria a tag, a release no GitHub e a marca numerada nas imagens (veja o [ADR-0023](docs/adr/0023-a-release-names-a-milestone-already-in-production.md)). A primeira release, a 0.1.0, ainda está por vir.

Um deploy reinicia a única instância da API. A API termina as requisições em andamento antes de sair (veja Shutdown em [`CODING_STANDARDS.md`](CODING_STANDARDS.md)), então o arquivo do Compose dá a ela um `stop_grace_period` de 30 segundos, e o Caddy segura as requisições novas por até 30 segundos (`lb_try_duration`) até a nova instância aceitar conexões.

Antes de a API subir, um serviço de execução única do Compose aplica as migrations pendentes do banco, e a API só sobe se ele terminar bem. Quando há migration pendente, o deploy gera um `pg_dump` antes, porque as migrations não têm caminho de volta (veja o [ADR-0015](docs/adr/0015-hand-written-sql-checked-by-sqlc-and-forward-only-migrations.md)).

O worker é um segundo serviço da mesma imagem, iniciado com o comando `worker`. Em um deploy ele para de pegar trabalhos e dá tempo para os que estão rodando terminarem. Um trabalho interrompido roda de novo depois (veja o [ADR-0016](docs/adr/0016-background-jobs-in-postgresql-run-by-a-worker-command.md)).

Os comandos estão _A definir_ e serão escritos com o primeiro código.

## Backup e manutenção

### Backup

| O quê | Como | Frequência | Retenção |
| --- | --- | --- | --- |
| Banco de dados (PostgreSQL) | pgBackRest, cifrado, em um bucket no Backblaze B2 | WAL continuamente, um backup diferencial por dia e um completo por semana | 30 dias, com restauração para qualquer momento |
| Banco de dados, antes de uma migration | `pg_dump` gerado pela publicação, guardado no servidor | Cada publicação com migration pendente | Os três últimos |
| Arquivos (Backblaze B2) | Versionamento de objetos no bucket de produção | A cada mudança | Uma versão antiga por 30 dias |

Os backups ficam em um provedor diferente do servidor. A chave que cifra os valores das credenciais e a senha dos backups nunca são guardadas junto com eles: são segredos do Environment do GitHub, com uma cópia no gerenciador de senhas do mantenedor. Um monitor fora do servidor avisa quando um backup não dá sinal (veja o [ADR-0020](docs/adr/0020-postgresql-backed-up-by-pgbackrest-and-production-files-in-backblaze-b2.md)).

### Restauração

Todo mês um workflow agendado cria um servidor, restaura nele o último backup, confere o resultado e destrói o servidor.

O passo a passo para restaurar à mão está _A definir_ e será escrito com a rotina de backup.

<!-- Passo a passo para restaurar o banco e os arquivos a partir de um backup. -->

### Manutenção

- **Dependências.** `task outdated` lista o que tem versão mais nova. Um workflow semanal roda esse comando e mantém a lista em uma única issue no GitHub, que avisa quem mantém o sistema; ele não abre pull request. `task outdated:machine`, rodado à mão, faz o mesmo para as ferramentas instaladas na máquina de quem mantém. Cada atualização é um pull request próprio (veja Versions em [`CODING_STANDARDS.md`](CODING_STANDARDS.md)).
- **Vulnerabilidades.** Um workflow semanal roda govulncheck, `pnpm audit` e Trivy nas imagens.
- **Logs.** Em produção, a saída dos containers é lida no servidor.

O restante está _A definir._

<!-- Rotinas recorrentes ainda por descrever: rotação de credenciais, atualização dos servidores. -->

## Documentação

- `docs/adr/`: decisões de arquitetura (ADRs).
- `GLOSSARY.md`: glossário dos termos do domínio.
- [`CODING_STANDARDS.md`](CODING_STANDARDS.md): como o código é escrito, para os agentes e para a revisão.
- [`DESIGN.md`](DESIGN.md): o design system: suas regras, variáveis de estilo, componentes e receitas.
- [`docs/providers.md`](docs/providers.md): os serviços externos de que o sistema depende, com o motivo, o custo e o que está configurado em cada um.
- [`docs/design/`](docs/design/): análise das referências visuais que embasam o design system.
- [`docs/agents/`](docs/agents/): configuração usada pelos agentes de IA neste repositório.

`docs/adr/` e `GLOSSARY.md` são criados conforme o domínio e a arquitetura forem definidos. O `CHANGELOG.md`, com o que mudou em cada release, é criado com a release 0.1.0.

## Licença

[GNU Affero General Public License, versão 3](LICENSE). Veja o [ADR-0021](docs/adr/0021-a-public-repository-under-agplv3.md).
