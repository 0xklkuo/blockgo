FROM golang:1.24.10 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -trimpath -o /out/blockgo ./cmd/blockgo
RUN go build -trimpath -o /out/blockgo-node ./cmd/blockgo-node

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=build /out/blockgo /usr/local/bin/blockgo
COPY --from=build /out/blockgo-node /usr/local/bin/blockgo-node
COPY configs /app/configs

ENTRYPOINT ["blockgo-node"]
CMD ["-version"]
