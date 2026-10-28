FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/biblioteca ./cmd/biblioteca

FROM alpine:3.22
RUN apk add --no-cache tzdata
ENV TZ=America/Sao_Paulo
COPY --from=build /out/biblioteca /usr/local/bin/biblioteca
ENTRYPOINT ["biblioteca"]
