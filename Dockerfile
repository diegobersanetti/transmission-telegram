FROM golang:alpine as build

RUN apk add --no-cache git

WORKDIR /src
COPY . .

RUN go build -o /transmission-telegram ./cmd/bot/

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=build /transmission-telegram /transmission-telegram

ENTRYPOINT ["/transmission-telegram"]
