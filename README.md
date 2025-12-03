# Gobid - Leilões em Tempo Real com Go

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)

## 🚀 Sobre o Projeto

**Gobid** é uma aplicação backend para uma plataforma de leilões, desenvolvida como projeto de estudo durante o curso de Go da [Rocketseat](https://www.rocketseat.com.br/).

O objetivo principal foi aplicar conceitos avançados de desenvolvimento backend em Go, construindo uma API robusta, organizada e performática, desde a interação com o banco de dados até a comunicação em tempo real com o cliente.

## ✨ Principais Aprendizados

Este projeto foi uma jornada de aprendizado sobre como construir aplicações Go prontas para produção. Os principais conceitos aplicados foram:

### 1. Arquitetura em Camadas (Layered Architecture)
O projeto é organizado em uma arquitetura limpa, separando as responsabilidades em camadas distintas:
- **`handlers`**: Camada de API, responsável por receber as requisições HTTP, validar dados de entrada e chamar os serviços correspondentes.
- **`services`**: Camada de serviço, onde reside a lógica de negócio da aplicação.
- **`usecase`**: Camada de casos de uso, que orquestra as regras de negócio mais complexas.
- **`store`**: Camada de acesso a dados, responsável pela comunicação com o banco de dados.

Essa separação facilita a manutenção, a testabilidade e a evolução do código.

### 2. Injeção de Dependência (Dependency Injection)
O código faz uso extensivo de injeção de dependência para desacoplar os componentes. Em vez de uma camada criar suas próprias dependências (por exemplo, um serviço instanciando o seu próprio repositório de dados), elas são "injetadas" de fora, geralmente no momento da inicialização da aplicação (no `main.go`).

**Vantagens:**
- **Testabilidade:** Facilita a criação de testes unitários, pois permite substituir dependências reais por implementações falsas (`mocks` ou `stubs`).
- **Flexibilidade:** Torna o código mais modular e fácil de reconfigurar ou estender.
- **Clareza:** As dependências de cada componente ficam explícitas em sua assinatura.

### 3. Geração de Código com `sqlc`
Uma das partes mais interessantes do projeto foi o uso do **`sqlc`**. Em vez de usar um ORM completo, escrevemos queries SQL puras e o `sqlc` gerou o código Go correspondente, totalmente type-safe.

**Vantagens:**
- **Performance:** Execução de SQL nativo.
- **Segurança:** Prevenção de SQL Injection, pois o `sqlc` cria funções tipadas.
- **Produtividade:** Geração automática do código de acesso a dados, evitando boilerplate.

### 4. Operações CRUD Completas
A API implementa todas as operações de **C**reate, **R**ead, **U**pdate e **D**elete para as principais entidades da aplicação, como `Usuários`, `Produtos` e `Leilões`. Isso solidificou o conhecimento sobre como construir APIs RESTful de forma eficiente.

### 5. Lances em Tempo Real com WebSockets
Para a funcionalidade de lances, o plano de estudo incluiu a implementação de **WebSockets**. Isso permite que o backend envie atualizações de novos lances para todos os clientes conectados em tempo real, sem a necessidade de o cliente ficar fazendo requisições (polling) a todo momento. É a tecnologia ideal para aplicações dinâmicas e interativas como um leilão.

### 6. Migrations de Banco de Dados com `tern`
O versionamento e a evolução do schema do banco de dados foram gerenciados com a ferramenta de migrations `tern`. Isso garante que as alterações no banco de dados sejam consistentes e reproduzíveis em qualquer ambiente.

### 7. Conscientização sobre Segurança: CSRF
Embora não tenha sido implementado (pois o foco era uma API que poderia ser consumida por um cliente mobile ou SPA com autenticação via token), o projeto foi uma oportunidade para aprender sobre ataques de **Cross-Site Request Forgery (CSRF)**.

**CSRF** é um tipo de ataque que engana o usuário autenticado a executar ações indesejadas. A proteção geralmente envolve o uso de tokens anti-CSRF, que garantem que a requisição foi originada pela própria aplicação, e não por um site malicioso.

## 🛠️ Tecnologias Utilizadas

- **Linguagem:** [Go](https://go.dev/)
- **Banco de Dados:** [PostgreSQL](https://www.postgresql.org/)
- **Geração de Código DB:** [sqlc](https://sqlc.dev/)
- **Migrations:** [tern](https://github.com/jackc/tern)
- **Containerização:** [Docker](https://www.docker.com/) e [Docker Compose](https://docs.docker.com/compose/)
- **Roteador HTTP:** `chi` (ou similar, como `gorilla/mux`)
- **Variáveis de Ambiente:** `godotenv`

## ⚙️ Como Executar o Projeto

1.  **Clone o repositório:**
    ```bash
    git clone https://github.com/seu-usuario/gobid.git
    cd gobid
    ```

2.  **Configure as variáveis de ambiente:**
    Renomeie o arquivo `.env.example` para `.env` e preencha com as informações do seu banco de dados PostgreSQL.

    ```env
    POSTGRES_USER=seu_usuario
    POSTGRES_PASSWORD=sua_senha
    POSTGRES_DB=gobid_db
    DB_SOURCE="postgresql://seu_usuario:sua_senha@localhost:5432/gobid_db?sslmode=disable"
    ```

3.  **Inicie os containers com Docker Compose:**
    Este comando irá iniciar a aplicação Go e o banco de dados PostgreSQL.

    ```bash
    docker-compose up --build
    ```

4.  **Acesse a API:**
    A aplicação estará disponível em `http://localhost:8080`.
