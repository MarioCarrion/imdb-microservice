FROM golang:1.23.4-bookworm AS builder

WORKDIR /build/

COPY . .

ENV CGO_ENABLED=1 \
    GOOS=linux

RUN go mod download

RUN go build -a -installsuffix cgo -ldflags "-extldflags -static" \
		github.com/MarioCarrion/todo-api/cmd/rest-server

#-

FROM debian:bookworm-20241223-slim

ENV DEBIAN_FRONTEND=noninteractive

RUN set -x && \
    apt-get update && \
    apt-get install -y \
      ca-certificates && \
      rm -rf /var/lib/apt/lists/*

WORKDIR /api/
ENV PATH=/api/bin/:$PATH

COPY --from=builder /build/rest-server ./bin/rest-server
COPY --from=builder /build/env.example .

EXPOSE 9234

CMD ["rest-server", "-env", "/api/env.example"]
