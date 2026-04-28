FROM golang:1.25-alpine3.23 AS builder
WORKDIR /build
COPY ./go.mod . 
RUN go mod download
COPY . .
RUN go build -o main cmd/app/main.go

FROM alpine:3.23.3
RUN apk add --no-cache tzdata
COPY ./configs /configs
COPY --from=builder /build/main /bin/main
ENTRYPOINT ["/bin/main"]