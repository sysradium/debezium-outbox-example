ARG SERVICE

FROM golang:1.22 AS builder
ARG SERVICE

WORKDIR /usr/src/app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN cd ${SERVICE} && make build

FROM alpine:latest
ARG SERVICE

WORKDIR /usr/src/app

COPY --from=builder /usr/src/app/${SERVICE}/bin/* .

CMD ["./main"]
