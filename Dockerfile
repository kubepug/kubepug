FROM golang:1.15.6-alpine AS builder

RUN apk add --update --no-cache ca-certificates bash make gcc musl-dev git openssh wget curl

WORKDIR /go/src/kubepug

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o ./kubepug cmd/kubepug.go

################

FROM alpine:3.12.1

RUN apk --no-cache add ca-certificates

COPY --from=builder /go/src/kubepug/bin/linux-amd64/kubepug /usr/local/bin/

ENTRYPOINT ["kubepug"]
