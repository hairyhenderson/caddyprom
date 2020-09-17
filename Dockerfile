FROM golang:1.14.3-alpine3.11 AS BUILD
RUN apk add -U --no-cache ca-certificates git gcc musl-dev

RUN go get -u github.com/caddyserver/xcaddy/cmd/xcaddy

WORKDIR /app/caddyprom

#cache dependencies
ADD /go.mod /app/caddyprom/
ADD /go.sum /app/caddyprom/
RUN go get github.com/lucaslorentz/caddy-docker-proxy/plugin/v2
RUN go mod download
RUN go get -v github.com/caddyserver/caddy/v2
RUN go get -v github.com/lucaslorentz/caddy-docker-proxy/plugin/v2

ADD /caddyprom.go /app/caddyprom/
ADD /caddyprom_test.go /app/caddyprom/
ADD /promhttp_fork_delegator.go /app/caddyprom/
ADD /promhttp_fork_instrument_server.go /app/caddyprom/
ADD /setup.go /app/caddyprom/

# RUN CGO_ENABLED=0 \
RUN xcaddy build \
    --output /app/caddy \
    --with github.com/lucaslorentz/caddy-docker-proxy/plugin/v2 \
    --with github.com/stutzlab/caddyprom=/app/caddyprom
    # --with github.com/miekg/caddy-prometheus=/tmp/caddy-prometheus


FROM alpine:3.12.0

EXPOSE 80 443 2019
ENV XDG_CONFIG_HOME /config
ENV XDG_DATA_HOME /data

WORKDIR /
COPY --from=BUILD /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=BUILD /app/caddy /bin/caddy

ENTRYPOINT ["/bin/caddy"]

CMD ["docker-proxy"]

