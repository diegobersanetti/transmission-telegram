FROM --platform=$BUILDPLATFORM golang:alpine AS build

ARG TARGETOS TARGETARCH TARGETVARIANT
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#v} \
    go build -ldflags="-s -w" -o /transmission-telegram ./cmd/bot/

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
COPY --from=build /transmission-telegram /transmission-telegram

USER 1000:1000

ENTRYPOINT ["/transmission-telegram"]
