# Mirrors vueblog-java/Dockerfile, which packaged the Spring Boot jar.
FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/vueblog ./cmd/vueblog

FROM alpine:3.20
RUN apk add --no-cache tzdata ca-certificates
ENV TZ=Asia/Shanghai
COPY --from=build /out/vueblog /app/vueblog
COPY resources /app/resources
EXPOSE 8081
ENTRYPOINT ["/app/vueblog"]
