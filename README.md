# "ToDo API" Microservice Example

[![codecov](https://codecov.io/gh/MarioCarrion/todo-api-microservice-example/branch/main/graph/badge.svg)](https://codecov.io/gh/MarioCarrion/todo-api-microservice-example)

## Introduction

Welcome! 👋

This is an educational repository that includes a microservice written in Go. It is used as the principal example of my video series: [Building Microservices in Go](https://www.youtube.com/playlist?list=PL7yAAGMOat_Fn8sAXIk0WyBfK_sT1pohu).

This repository **is not** a template **nor** a framework, it's a **collection of patterns and guidelines** I've successfully used to deliver enterprise microservices when using Go, and just like with everything in Software Development some trade-offs were made.

My end goal with this project is to help you learn another way to structure your Go project with 3 final goals:

1. It is _enterprise_, meant to last for years,
1. It allows a team to collaborate efficiently with little friction, and
1. It is as idiomatic as possible.

Join the fun at [https://youtube.com/MarioCarrion](https://www.youtube.com/c/MarioCarrion).

## Domain Driven Design

This project uses a lot of the ideas introduced by Eric Evans in his book [Domain Driven Design](https://www.domainlanguage.com/), I do encourage reading that book but before I think reading [Domain-Driven Design Distilled](https://smile.amazon.com/Domain-Driven-Design-Distilled-Vaughn-Vernon/dp/0134434420/) makes more sense, also there's a free to download [DDD Reference](https://www.domainlanguage.com/ddd/reference/) available as well.

On YouTube I created [a playlist](https://www.youtube.com/playlist?list=PL7yAAGMOat_GJqfTdM9PBdTRSH7jXs6mI) that includes some of my favorite talks and webinars, feel free to explore that as well.

## Project Structure

Talking specifically about microservices **only**, the structure I like to recommend is the following, everything using `<` and `>` depends on the domain being implemented and the bounded context being defined.

- [ ] `build/`: defines the code used for creating infrastructure as well as docker containers.
  - [ ] `<cloud-providers>/`: define concrete cloud provider.
  - [ ] `<executableN>/`: contains a Dockerfile used for building the binary.
- [ ] `cmd/`
  - [ ] `<primary-server>/`: uses primary database.
  - [ ] `<replica-server>/`: uses readonly databases.
  - [ ] `<binaryN>/`
- [X] `db/`
  - [X] `migrations/`: contains database migrations.
  - [ ] `seeds/`: contains file meant to populate basic database values.
- [ ] `internal/`: defines the _core domain_.
  - [ ] `<datastoreN>/`: a concrete _repository_ used by the domain, for example `postgresql`
  - [ ] `http/`: defines HTTP Handlers.
  - [ ] `service/`: orchestrates use cases and manages transactions.
- [X] `pkg/` public API meant to be imported by other Go package.

There are cases where requiring a new bounded context is needed, in those cases the recommendation would be to
define a package like `internal/<bounded-context>` that then should follow the same structure, for example:

* `internal/<bounded-context>/`
  * `internal/<bounded-context>/<datastoreN>`
  * `internal/<bounded-context>/http`
  * `internal/<bounded-context>/service`

## Tools

Please refer to the documentation in [internal/tools/](internal/tools/README.md).

## Features

Icons meaning:

* <img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video"> means a link to Youtube video.
* <img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/link.svg" width="20" height="20" alt="Blog post"> means a link to Blog post.

In no particular order:

- [X] Project Layout [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/LUvid5TJ81Y) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/link.svg" width="20" height="20" alt="Blog post">](https://mariocarrion.com/2021/03/21/golang-microservices-domain-driven-design-project-layout.html)
- [X] Dependency Injection [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/Z89UU4vSayY)
- [X] [Secure Configuration](docs/SECURE\_CONFIGURATION.md)
  - [X] Using [Hashicorp Vault](https://www.hashicorp.com/products/vault) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/7UmJR0dOkjM)
  - [X] Using [AWS SSM](https://aws.amazon.com/systems-manager/features/#Parameter_Store) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/link.svg" width="20" height="20" alt="Blog post">](https://mariocarrion.com/2019/11/20/golang-aws-ssm-env-vars-package.html)
- [ ] Infrastructure as code
- [X] [Metrics, Traces and Logging using OpenTelemetry](docs/METRICS\_TRACES\_LOGGING.md) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/bytCFQJ43DE)
- [ ] Caching
  - [X] Memcached [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/yKI-sy70PwA) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/link.svg" width="20" height="20" alt="Blog post">](https://mariocarrion.com/2021/01/30/tips-building-microservices-in-go-golang-caching-memcached.html)
  - [ ] Redis [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/wj6-w0DLKRw)
- [X] Persistent Storage
  - [X] Repository Pattern [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/Z89UU4vSayY) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/link.svg" width="20" height="20" alt="Blog post">](https://mariocarrion.com/2021/04/04/golang-microservices-repository-pattern.html)
  - [X] Database migrations [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/EavdaeUmn64) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/link.svg" width="20" height="20" alt="Blog post">](https://mariocarrion.com/2021/01/10/golang-tools-for-database-schema-migrations.html)
  - [ ] MySQL
  - [X] [PostgreSQL](docs/PERSISTENT\_STORAGE.md) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/link.svg" width="20" height="20" alt="Blog post">](https://mariocarrion.com/2021/03/02/tips-building-microservices-in-go-golang-databases-postgresql-conclusion.html)
      - [`jmoiron/sqlx`](https://github.com/jmoiron/sqlx), [`jackc/pgx`](https://github.com/jackc/pgx) and [`database/sql`](https://pkg.go.dev/database/sql) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/l8t6UKM1kro) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/link.svg" width="20" height="20" alt="Blog post">](https://mariocarrion.com/2021/02/05/tips-building-microservices-in-go-golang-databases-postgresql-part1.html)
      - [`go-gorm/gorm`](https://github.com/go-gorm/gorm) and [`volatiletech/sqlboiler`](https://github.com/volatiletech/sqlboiler) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/CT2v0Xas8Sc) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/link.svg" width="20" height="20" alt="Blog post">](https://mariocarrion.com/2021/02/10/tips-building-microservices-in-go-golang-databases-postgresql-gorm-orm.html)
      - [`Masterminds/squirrel`](https://github.com/Masterminds/squirrel) and [`kyleconroy/sqlc`](https://github.com/kyleconroy/sqlc) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/CT2v0Xas8Sc) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/link.svg" width="20" height="20" alt="Blog post">](https://mariocarrion.com/2021/02/23/tips-building-microservices-in-go-golang-databases-postgresql-sqlc-squirrel.html)
- [ ] REST APIs
  - [X] HTTP Handlers [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/CLdxwJCvTZE) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/link.svg" width="20" height="20" alt="Blog post">](https://mariocarrion.com/2021/04/18/golang-microservices-rest-api-http-handlers.html)
  - [X] Custom JSON Types [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/UmVYkEYm4hw) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/link.svg" width="20" height="20" alt="Blog post">](https://mariocarrion.com/2021/04/28/golang-microservices-rest-api-custom-json-type.html)
  - [ ] Versioning [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/4THy4iBQpFA)
  - [X] Error Handling [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/uQOfXL6IFmQ)
  - [X] [OpenAPI 3 and Swagger-UI](docs/OPENAPI3\_SWAGGER.md) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/HwtOAc0M08o) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/link.svg" width="20" height="20" alt="Blog post">](https://mariocarrion.com/2021/05/02/golang-microservices-rest-api-openapi3-swagger-ui.html)
  - [ ] Authorization
- [ ] Events and Messaging
  - [ ] [Apache Kafka](https://kafka.apache.org/) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/jr7OULxYm0A)
  - [ ] [RabbitMQ](https://www.rabbitmq.com/) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/L0yJxCKrkIY)
  - [ ] [Redis](https://redis.io/) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/uzuwZNiN4Y8)
- [ ] Testing
  - [X] Type-safe mocks with [`maxbrunsfeld/counterfeiter`](https://github.com/maxbrunsfeld/counterfeiter) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/ENqwq64TsDk) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/link.svg" width="20" height="20" alt="Blog post">](https://mariocarrion.com/2019/06/24/golang-tools-counterfeiter.html)
  - [X] Equality with [`google/go-cmp`](https://github.com/google/go-cmp) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/ae15DzSwNnU) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/link.svg" width="20" height="20" alt="Blog post">](https://mariocarrion.com/2021/01/22/go-package-equality-google-go-cmp.html)
  - [X] Integration tests for Datastores with [`ory/dockertest`](https://github.com/ory/dockertest) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/a-CCceqerhg) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/link.svg" width="20" height="20" alt="Blog post">](https://mariocarrion.com/2021/03/14/golang-package-testing-datastores-ory-dockertest.html)
  - [X] REST APIs [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/lMrWO7OUMdY) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/link.svg" width="20" height="20" alt="Blog post">](https://mariocarrion.com/2021/04/25/golang-microservices-rest-api-testing.html)
- [ ] Containerization using Docker [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/u_ayzie9pAQ)
- [ ] Graceful Shutdown [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/VXxe7-b5euo)
- [ ] Search Engine using [ElasticSearch](https://www.elastic.co/elasticsearch/) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/ZrdbQRYst5E)
- [ ] Documentation
  - [C4 Model](https://c4model.com/) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/pZ2z2tZkMsE)
- [ ] Cloud Design Patterns
  * Reliability
    - [ ] [Circuit Breaker](https://en.wikipedia.org/wiki/Circuit_breaker_design_pattern) [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/UnL2iGcD7vE)
- [ ] Tools as Dependencies [<img src="https://github.com/MarioCarrion/MarioCarrion/blob/main/youtube.svg" width="20" height="20" alt="YouTube video">](https://youtu.be/g_5n0W27XcY)
- [ ] Whatever else I forgot to include

## More ideas

* [2016: Peter Bourgon's: Repository structure](https://peter.bourgon.org/go-best-practices-2016/#repository-structure)
* [2016: Ben Johnson's: Standard Package Layout](https://medium.com/@benbjohnson/standard-package-layout-7cdbc8391fc1)
* [2017: William Kennedy's: Design Philosophy On Packaging](https://www.ardanlabs.com/blog/2017/02/design-philosophy-on-packaging.html)
* [2017: Jaana Dogan's: Style guideline for Go packages](https://rakyll.org/style-packages/)
* [2018: Kat Zien - How Do You Structure Your Go Apps](https://www.youtube.com/watch?v=oL6JBUk6tj0)

## Running project locally using Docker Compose

Originally added as part of [Building Microservices In Go: Containerization with Docker](https://youtu.be/u_ayzie9pAQ), `docker compose` has evolved and with it the way to run everything locally. Make sure you are running a recent version of Docker Compose. The configuration in this repository and the instructions below are known to work:

* Engine: **27.4.0**, and
* Compose: **v2.31.0-desktop.2**

This project takes advantage of [Go's build constrains](https://pkg.go.dev/go/build) and [Docker's arguments](https://docs.docker.com/reference/dockerfile/#arg) to build and run the [rest-server](cmd/rest-server) using any of the following types of message broker:

* RabbitMQ
* Kafka

The `docker compose` instructions are executed in the form of:

```
docker compose -f compose.yml -f compose.<type>.yml <command>
```

Where:

* `<type>`: Indicates what message broker to use. Currently there are two supported values: `rabbitmq` and `kafka`; those effectively match the compose filename itself.
* `<command>`: Indicates the docker compose command to use.

For example to build images using RabbitMQ as the message broker you would do:

```
docker compose -f compose.yml -f compose.rabbitmq.yml build
```

Then to start the containers you would do:

```
docker compose -f compose.yml -f compose.rabbitmq.yml up
```

Once you `up` all the containers you can access the Swagger UI at http://127.0.0.1:9234/static/swagger-ui/ .

## Diagrams

To start a local HTTP server that serves a graphical editor:

```
mdl serve github.com/MarioCarrion/todo-api/internal/doc -dir docs/diagrams/
```

To generate JSON artifact for uploading to [structurizr](https://structurizr.com/):

```
stz gen github.com/MarioCarrion/todo-api/internal/doc
```
