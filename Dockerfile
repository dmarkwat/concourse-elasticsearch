FROM golang:1.14.2-buster

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . ./
RUN make build

FROM debian:buster
COPY --from=0 /build/build/* /opt/resource/
