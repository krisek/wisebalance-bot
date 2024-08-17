FROM golang:1.22 as builder
COPY main.go /main.go
ENV GOROOT=/usr/local/go
RUN CGO_ENABLED=0 go build -o /wisebalance-bot /main.go

FROM gcr.io/distroless/base
COPY --from=builder /wisebalance-bot /wisebalance-bot
USER 65534:65534
ENTRYPOINT ["/wisebalance-bot"]
