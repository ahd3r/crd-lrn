FROM golang:1.25.3 AS builder
WORKDIR /workspace
COPY ./go.mod ./
COPY ./main.go ./
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /ns-controller main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /ns-controller .
USER 65532:65532

ENTRYPOINT ["/ns-controller"]
