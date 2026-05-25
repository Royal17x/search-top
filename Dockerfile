# build
FROM golang:1.22-alpine as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/server ./cmd/server

# start
FROM scratch
COPY --from=builder /bin/server /server
ENTRYPOINT ["/server"]

