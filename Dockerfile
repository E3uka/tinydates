FROM golang:1.22-alpine AS builder

ENV GO111MODULE=on

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

RUN go build cmd/main.go

FROM alpine:latest

WORKDIR /bin
COPY --from=builder /go/src/app/main /bin
COPY --from=builder /go/src/app/.env /bin

EXPOSE 8080

CMD ["./main"]
