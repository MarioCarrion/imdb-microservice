FROM golang:1.23.5-bookworm AS builder

WORKDIR /build/

COPY . .

RUN go mod download

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags "-extldflags -static" \
      -o elasticsearch-indexer github.com/MarioCarrion/todo-api/cmd/elasticsearch-indexer-redis

#-

FROM debian:bookworm-20250113-slim

RUN set -x && \
    apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
      ca-certificates && \
      rm -rf /var/lib/apt/lists/*

WORKDIR /api/
ENV PATH=/api/bin/:$PATH

COPY --from=builder /build/elasticsearch-indexer ./bin/elasticsearch-indexer
COPY --from=builder /build/env.example .

CMD ["elasticsearch-indexer", "-env", "/api/env.example"]
