FROM golang:latest
RUN mkdir -p /app
RUN go install github.com/pressly/goose/cmd/goose@latest
COPY . /app
WORKDIR /app
RUN go build
## Moved to docker-compose
# CMD ["/app/go_rss_demo"]