# vueblog: Java → Go migration notes

Source: `MarkerHub/vueblog` @ `c042ee4a37e865c08addc680af26428de7e8ed89` (`vueblog-java` module only)
Licence: Apache-2.0 (carried over verbatim in `LICENSE`)

The Vue frontend (`vueblog-vue/`) is untouched — it talks to this service over
the same HTTP contract.

## File mapping

| Java | Go |
|---|---|
| `VueblogApplication.java` | `cmd/vueblog/main.go` |
| `common/lang/Result.java` | `internal/common/lang/result.go` |
| `common/dto/LoginDto.java` | `internal/common/dto/login_dto.go` |
| `common/exception/GlobalExceptionHandler.java` | `internal/common/exception/global_exception_handler.go` |
| `config/CorsConfig.java` | `internal/config/cors.go` |
| `config/ShiroConfig.java` | `internal/config/shiro_config.go` + `internal/config/router.go` |
| `config/MybatisPlusConfig.java` | `internal/config/mybatis_plus_config.go` |
| `controller/AccountController.java` | `internal/controller/account_controller.go` |
| `controller/BlogController.java` | `internal/controller/blog_controller.go` |
| `controller/UserController.java` | `internal/controller/user_controller.go` |
| `entity/Blog.java`, `entity/User.java` | `internal/entity/blog.go`, `user.go` |
| `mapper/*.java` + `mapper/*.xml` | `internal/mapper/` |
| `service/`, `service/impl/` | `internal/service/service.go` |
| `shiro/AccountRealm.java` | `internal/shiro/account_realm.go` |
| `shiro/JwtFilter.java` | `internal/shiro/jwt_filter.go` |
| `shiro/JwtToken.java`, `AccountProfile.java` | `internal/shiro/subject.go` |
| `util/JwtUtils.java` | `internal/util/jwt_utils.go` |
| `util/ShiroUtil.java` | `internal/shiro/shiro_util.go` (see Deviations) |
| `VueblogApplicationTests.java` | `test/vueblog_application_test.go` |
| `application.yml` | `internal/config/properties.go` + `resources/application.yml` |

Not migrated: `CodeGenerator.java` is a MyBatis-Plus scaffolding script that
generates Java sources at development time. It produces no runtime behaviour and
has no meaning in Go.

## Dependency substitutions

| Java | Go |
|---|---|
| Spring Boot Web / DispatcherServlet | `net/http` + `ServeMux` (Go 1.22 method patterns) |
| Apache Shiro | hand-written `internal/shiro` (Subject, realm, authenticating filter) |
| MyBatis-Plus | hand-written `internal/mybatisplus` + `database/sql` |
| jjwt 0.9.1 | `github.com/golang-jwt/jwt/v5` |
| Lombok | plain structs and explicit methods |
| Jackson | `encoding/json` + custom marshalers for the date types |
| hutool `SecureUtil.md5` | `crypto/md5` |
| hutool `BeanUtil.copyProperties` | explicit field copies |
| javax.validation | `internal/validation` |
| mysql-connector-java | `github.com/go-sql-driver/mysql` |

Shiro binds the Subject to a `ThreadLocal`; the Go port carries it on the
request `context`, which is the idiomatic equivalent and behaves identically
because the Subject's lifetime is exactly one request.

## Behaviour preserved deliberately — including the ugly parts

These are the details most likely to be "fixed" by accident. Each is reproduced
and pinned by a test.

**1. The JWT secret is base64-decoded, not used raw.**
jjwt 0.9.1's `signWith(SignatureAlgorithm, String)` calls
`TextCodec.BASE64.decode(secret)` before using the bytes as the HMAC key, and
`setSigningKey(String)` does the same. So `f4e2e52034348f86b67cde581c0f9eb5`
becomes a 24-byte key, not 32 ASCII bytes. Using the raw string would produce
tokens the Java service rejects and vice versa — the two deployments would not
interoperate. `TestSecretIsBase64Decoded`.

**2. A wrong password returns HTTP 200.**
`AccountController.login` *returns* `Result.fail("密码不正确")` instead of
throwing, so the status stays 200 with `code: 400` in the body. Only the
`Assert.notNull` path (unknown user) produces a real 400.
`TestLoginWrongPassword`.

**3. Realm failures return HTTP 200 too.**
`JwtFilter.onLoginFailure` writes JSON straight to the response and never calls
`setStatus`. So "账户不存在" and "账户已被锁定" arrive as HTTP 200 carrying
`code: 400`, unlike every other error.
`TestValidTokenForMissingUserWritesBodyWith200`, `TestLockedAccountWritesBodyWith200`.

**4. Editing a non-existent blog id throws NPE with a null message.**
`BlogController.edit` does `temp = blogService.getById(id)` then
`temp.getUserId()` with no null check. The project targets Java 1.8
(`<java.version>1.8</java.version>`), which predates helpful NPE messages, so
`getMessage()` is null and `Result.fail(null)` serialises `"msg": null`. This is
a genuine defect in the original; it is preserved, not repaired.
`TestBlogEditMissingIdThrowsNPEWithNullMessage`.

**5. Logout does not invalidate the token.**
Authentication is stateless JWT. `SecurityUtils.getSubject().logout()` tears down
the Shiro session but cannot revoke an already-issued token, so the same token
keeps working until it expires. `TestLogoutDoesNotInvalidateTheToken`.

**6. `/user/index` ignores who is logged in.**
It is hardcoded to `userService.getById(1L)`. Authenticating as user 2 still
returns user 1. `TestUserIndexReturnsUserOne`.

**7. `/user/save` persists nothing.**
It validates and echoes the payload back. There is no save call in the original.
`TestUserSaveIsPublic`.

**8. An absent token is not rejected by the filter.**
`onAccessDenied` returns `true` when the `Authorization` header is missing, so
anonymous requests proceed. Only `@RequiresAuthentication` on the individual
handler stops them — which is what makes `/blogs` and `/blog/{id}` public.

**9. `copyProperties` ignores four fields.**
`BeanUtil.copyProperties(blog, temp, "id","userId","created","status")` copies
everything *except* those, so a client cannot rewrite ownership, timestamps or
status through `/blog/edit`. `TestBlogEditUpdate` asserts that a request sending
`userId: 999, status: 42` changes neither.

**10. Two different date formats.**
`Blog.created` carries `@JsonFormat(pattern="yyyy-MM-dd")` → `"2020-05-21"`.
`User.created` has no annotation, and Spring Boot disables
`WRITE_DATES_AS_TIMESTAMPS`, so it is ISO-8601 without an offset →
`"2020-04-20T10:44:01"`. `TestBlogCreatedIsDateOnly`, `TestUserCreatedIsIsoDateTime`.

**11. Jackson's ALWAYS inclusion.**
This app never sets `serializationInclusion`, so null fields are emitted:
`"email": null`, `"data": null`. `TestNullDatesSerialiseAsNull`.

**12. CORS header precedence.**
`JwtFilter.preHandle` is a servlet filter and runs *before* DispatcherServlet, so
an OPTIONS preflight is answered with the filter's values
(`GET,POST,OPTIONS,PUT,DELETE`) and never reaches `CorsConfig`
(`GET,HEAD,POST,PUT,DELETE,OPTIONS`). `router.go` reproduces that layering.
`TestOptionsShortCircuits`.

## Deviations

**`ShiroUtil` lives in `internal/shiro`, not `internal/util`.**
The Java class is `com.markerhub.util.ShiroUtil`, but `util` already holds
`JwtUtils`, which `shiro` imports. Keeping `ShiroUtil` in `util` would create the
import cycle `util → shiro → util`, which Go forbids. Behaviour is unchanged.
This is the only structural departure from the Java package layout.

**Login response key order.**
Java builds a `HashMap` via hutool's `MapUtil.builder()`, whose JSON key order is
unspecified. Go uses a struct, giving deterministic `id, username, avatar, email`
order. JSON object key order is not semantically significant.

**Redis is not wired.**
`shiro-redis` backs Shiro's session store and auth cache. Because authentication
is stateless JWT and every request re-authenticates from the token, the cache
never changes an authorisation outcome — it is a performance layer. Its only
observable effect, session teardown on logout, is already a no-op against a
stateless token (see #5). The config values are parsed and retained in
`internal/config/properties.go` so the wiring can be added without a redesign.

## Verification status — read this before trusting byte-equality

Everything below was actually run on this machine:

- `go build ./...` — clean
- `go vet ./...` — clean
- `gofmt -l .` — clean
- `go test ./... -count=1` — **66 tests, all passing**
- Coverage — **75.4%** of `./internal/...`
- The binary boots, serves, and returns correct CORS/404/error responses

**What was NOT verified, and why.** No JDK is installed on this machine, and the
original needs MySQL and Redis besides. The Java application was therefore never
run, so no output was compared side by side. Every fidelity claim above is
derived from reading the source and the pinned library versions, not from a
differential test.

The one area where that matters most is the **`Page` JSON shape**.
`BlogController.list` returns MyBatis-Plus's `Page` object directly inside
`Result.data`, so its field set is part of the public API. `internal/mybatisplus/page.go`
implements MP **3.2.0**'s shape — `records, total, size, current, ascs, descs,
optimizeCountSql, searchCount, pages`. The core fields (`records`, `total`,
`size`, `current`, `pages`) are certain; the auxiliary flags are reconstructed
from that version's getters and are the most likely place for a discrepancy,
since MP changed them across 3.x (later versions expose `orders` and `hitCount`
instead of `ascs`/`descs`). If you can run the Java app once, diff a real
`GET /blogs` response against this and adjust that struct — nothing else depends
on it.

## Running

```sh
make build            # compile
make test             # 66 tests, no database required
make cover            # coverage report
make run              # starts on :8081, same as server.port
```

The test suite substitutes in-memory mappers seeded from `resources/vueblog.sql`,
so it needs neither MySQL nor Redis. `make run` expects MySQL on
`localhost:3306/vueblog`; without it the service still starts and each
data-touching endpoint returns a 400 carrying the driver error, which is the
same shape the Java `RuntimeException` handler produces.
