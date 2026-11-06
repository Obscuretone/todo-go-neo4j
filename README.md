# Go / Neo4j task API

A small Go HTTP API using Gorilla Mux and the Neo4j Go driver. Tasks are stored
as graph nodes, with an optional parent ID for task relationships.

## Local development

The checked-in Compose setup starts Neo4j and a development container using
`gin` as a rebuild/restart wrapper. It uses disposable `neo4j/password`
credentials, not a production deployment configuration.

```sh
docker compose up --build
curl http://localhost:8080/tasks
```

The API listens on port 8080; Neo4j Browser is exposed on port 7474 and Bolt
on port 7687. The development proxy is exposed on port 3000. Live reloading
is experimental: the current bind mount targets `/app`, while the container
build directory is `/todo-app`. Rebuild the container after editing source.

## Routes

- `GET /tasks`
- `POST /tasks`
- `GET /tasks/{taskID}`
- `PUT /tasks/{taskID}`
- `DELETE /tasks/{taskID}`

A task contains `id`, `title`, `completed`, and optional `parent_id` fields.

## Build check

```sh
go build ./app/...
```

The driver currently uses the Compose hostname and development credentials
in `app/config/neo4j.go`; the Compose environment variables do not override
them. This is an ecosystem demo, with no claim of production readiness.
