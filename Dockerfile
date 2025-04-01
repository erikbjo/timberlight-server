FROM golang:1.24 AS builder

LABEL authors="erbj@stud.ntnu.no,simonhou@stud.ntnu.no"
LABEL stage=builder

WORKDIR /

RUN apt-get update && apt-get install -y libproj-dev pkg-config

COPY ./go.mod ./go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app ./cmd/api/main.go

FROM alpine:3.20

COPY --from=builder /app /app

CMD [ "/app" ]
