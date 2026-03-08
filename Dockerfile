FROM golang:1.24-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/universald ./cmd/universald

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=build /out/universald /app/universald
COPY web /app/web

ENV PORT=10000
ENV UNIVERSALD_DATA_DIR=/var/data
ENV UNIVERSALD_DB=/var/data/universald.db
ENV UNIVERSALD_UPLOADS=/var/data/panel_uploads
ENV UNIVERSALD_WEB=/app/web
ENV UNIVERSALD_OPEN=false
ENV UNIVERSALD_SAFE_MODE=true

EXPOSE 10000

CMD ["/app/universald"]
