FROM golang:1.18.1-alpine3.15 AS builder

COPY . /github.com/kosdirus/cryptool/
WORKDIR /github.com/kosdirus/cryptool/

RUN go mod download
RUN go build -o ./bin/app cmd/server/main.go


# Run stage / Deploy
FROM alpine AS production

WORKDIR /root/
COPY --from=0 /github.com/kosdirus/cryptool/bin/app .

EXPOSE 8080
#USER nonroot:nonroot
ENTRYPOINT ["./app"]