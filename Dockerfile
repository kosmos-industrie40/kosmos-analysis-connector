FROM golang:1.15-buster AS builder
COPY . /go/src/github.com/kosmos-industrie40/kosmos-analyse-connector
WORKDIR /go/src/github.com/kosmos-industrie40/kosmos-analyse-connector
RUN go build -o /usr/local/bin/connector src/main.go

FROM gcr.io/distroless/base-debian10:latest
COPY --from=builder /usr/local/bin/connector /usr/local/bin/connector
COPY exampleConfiguration.yaml ./exampleConfiguration.yaml
USER nonroot:nonroot

ENTRYPOINT ["/usr/local/bin/connector"]
