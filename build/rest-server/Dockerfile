FROM golang:1.23.5-bookworm AS builder

WORKDIR /build/

COPY . .

RUN go mod download

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags "-extldflags -static" \
      github.com/MarioCarrion/todo-api/cmd/rest-server

RUN CGO_ENABLED=1 GOOS=linux go install -a -installsuffix cgo -ldflags "-extldflags -static" \
      github.com/jackc/tern/v2@v2.1.1

#-

FROM debian:bookworm-20250113-slim

RUN set -x && \
    apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
      ca-certificates && \
      rm -rf /var/lib/apt/lists/*

WORKDIR /api/
ENV PATH=/api/bin/:$PATH

COPY --from=builder /build/rest-server ./bin/rest-server
COPY --from=builder /go/bin/tern ./bin/tern
COPY --from=builder /build/env.example .
COPY --from=builder /build/db/ .

EXPOSE 9234

CMD ["rest-server", "-env", "/api/env.example"]
