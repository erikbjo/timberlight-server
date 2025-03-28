FROM golang:1.24 AS builder

LABEL authors="erbj@stud.ntnu.no,simonhou@stud.ntnu.no"
LABEL stage=builder

WORKDIR /server

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o executable ./cmd/api/main.go

EXPOSE 8080

RUN touch .env

CMD ["./executable"]