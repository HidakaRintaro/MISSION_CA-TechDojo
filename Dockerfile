FROM golang:latest

RUN mkdir /app

WORKDIR /app

# mysql関連
RUN go get github.com/go-sql-driver/mysql