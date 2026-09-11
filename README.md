# vueblog-go

A Go port of the `vueblog-java` module from [MarkerHub/vueblog](https://github.com/MarkerHub/vueblog),
preserving its HTTP contract exactly. The original Vue frontend works against it
unchanged.

See [MIGRATION.md](MIGRATION.md) for the file-by-file mapping, the behaviours
preserved verbatim (including several genuine quirks of the original), and the
verification status.

## API

| Method | Path | Auth | Behaviour |
|---|---|---|---|
| `POST` | `/login` | — | Validates credentials, returns a JWT in the `Authorization` response header |
| `GET` | `/logout` | required | Clears the Shiro subject (does not revoke the token) |
| `GET` | `/blogs?currentPage=N` | — | Page of 5, newest first |
| `GET` | `/blog/{id}` | — | One blog, or 400 `该博客已被删除` |
| `POST` | `/blog/edit` | required | Creates or updates; only the owner may update |
| `GET` | `/user/index` | required | Returns user id 1 (hardcoded in the original) |
| `POST` | `/user/save` | — | Validates and echoes; persists nothing |

All responses use the `Result` envelope:

```json
{ "code": 200, "msg": "操作成功", "data": null }
```

Note that `code` and the HTTP status do not always agree — see MIGRATION.md
items 2 and 3.

## Build and test

```sh
make build     # go build ./...
make test      # 66 tests, no MySQL or Redis needed
make cover     # coverage over ./internal/...
make run       # listens on :8081
```

The suite swaps the mapper layer for in-memory implementations seeded from
`resources/vueblog.sql`, leaving the filter, realm, controllers, exception
handler and serialisation exactly as they run in production.

## Configuration

Defaults mirror the committed `application.yml`. Override with a file:

```sh
go run ./cmd/vueblog -config resources/application.yml
```

| Key | Default |
|---|---|
| `server.port` | `8081` |
| `markerhub.jwt.secret` | `f4e2e52034348f86b67cde581c0f9eb5` |
| `markerhub.jwt.expire` | `604800` (7 days) |
| `spring.datasource.url` | `root:admin@tcp(localhost:3306)/vueblog?...` |

The datasource URL uses Go driver DSN syntax rather than a JDBC URL. Load the
schema from `resources/vueblog.sql`.

## Licence

Apache-2.0, inherited from the upstream project.
