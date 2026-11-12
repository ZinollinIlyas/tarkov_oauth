FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go mod tidy && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/app /app/app
# config.json обязателен для этого сервиса; значения перекрываются env
# (postgres.host -> POSTGRES_HOST, jwt.secret -> JWT_SECRET и т.д.)
COPY config.json /app/config.json
EXPOSE 8080
ENTRYPOINT ["/app/app"]
