FROM golang:1.23-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-w -s -X github.com/chrisvanmeer/proxyglass/internal/build.Version=${VERSION}" -o /out/proxyglass ./cmd/proxyglass

FROM alpine:3.21
RUN apk add --no-cache ca-certificates && addgroup -S proxyglass && adduser -S -G proxyglass proxyglass \
    && mkdir -p /etc/proxyglass /var/lib/proxyglass /var/log/proxyglass \
    && chown -R proxyglass:proxyglass /etc/proxyglass /var/lib/proxyglass /var/log/proxyglass
COPY --from=build /out/proxyglass /usr/bin/proxyglass
USER proxyglass
EXPOSE 8080
ENTRYPOINT ["/usr/bin/proxyglass"]
CMD ["--config", "/etc/proxyglass/config.yml"]
