FROM golang:1.22.2-alpine3.19 AS builder
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o /build/bin/app

FROM scratch
COPY --from=builder /build/bin/app /bin/app
ADD ca-cert.pem /bin/app
ADD ca-key.pem /bin/app
EXPOSE 5432/udp
EXPOSE 4433/tcp
CMD ["./bin/app"]