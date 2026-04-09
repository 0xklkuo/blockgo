FROM golang:1.24.10 AS build

WORKDIR /src
COPY . .
RUN go mod tidy
RUN go build -o /out/blockgo ./cmd/blockgo
RUN go build -o /out/blockgo-node ./cmd/blockgo-node

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=build /out/blockgo /usr/local/bin/blockgo
COPY --from=build /out/blockgo-node /usr/local/bin/blockgo-node
COPY configs /app/configs
CMD ["blockgo-node", "-version"]
