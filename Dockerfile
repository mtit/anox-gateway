FROM golang:1.24 AS build
WORKDIR /src

ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/anox-gateway .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/anox-gateway /usr/local/bin/anox-gateway
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/anox-gateway"]