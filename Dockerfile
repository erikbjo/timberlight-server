FROM golang:1.24 AS builder

LABEL authors="erbj@stud.ntnu.no,simonhou@stud.ntnu.no"
LABEL stage=builder

WORKDIR /app

RUN apt-get update && apt-get install -y libproj-dev pkg-config

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /api ./cmd/api/main.go

FROM debian:bookworm-slim

WORKDIR /root/

# Install runtime dependencies
RUN apt-get update && apt-get install -y libproj-dev

# Copy the built binary from builder stage
COPY --from=builder /api /api
COPY --from=builder /app/proxy.json proxy.json
RUN touch .env

EXPOSE 8080

CMD [ "/api" ]
