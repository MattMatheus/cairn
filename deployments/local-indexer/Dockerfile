FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/cairn-indexer ./cmd/cairn-indexer

FROM alpine:3.22
RUN adduser -D -u 10001 cairn
USER cairn
WORKDIR /app
COPY --from=build /out/cairn-indexer /app/cairn-indexer
EXPOSE 8080
ENV CAIRN_WORKSPACE_ROOT=/workspace
ENV CAIRN_INDEXER_ADDR=:8080
ENTRYPOINT ["/app/cairn-indexer"]
