FROM golang:1.22 as builder
ENV GO111MODULE=on
ENV GOROOT=/usr/local/go
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy
RUN go get github.com/matrix-org/gomatrix@latest
COPY main.go ./
RUN CGO_ENABLED=0 go build -o /wisebalance-bot /app/main.go

FROM gcr.io/distroless/base
COPY --from=builder /wisebalance-bot /wisebalance-bot
USER 65534:65534
ENTRYPOINT ["/wisebalance-bot"]
