FROM golang:1.25 AS build

WORKDIR /srv/root

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o RealistikAPI \
    && chmod +x scripts/start.sh


FROM debian:bookworm-slim

# curl is for the compose healthcheck.
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /srv/root

COPY --from=build /srv/root/RealistikAPI ./
COPY --from=build /srv/root/scripts/start.sh ./scripts/start.sh

EXPOSE 80

CMD ["./scripts/start.sh"]
