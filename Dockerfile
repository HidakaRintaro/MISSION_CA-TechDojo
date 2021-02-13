FROM golang:latest

RUN mkdir /app

WORKDIR /app

RUN go get github.com/go-sql-driver/mysql
RUN go get github.com/gorilla/mux
RUN go get github.com/dgrijalva/jwt-go