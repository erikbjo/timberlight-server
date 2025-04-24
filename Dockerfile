FROM golang:1.24 AS builder

LABEL authors="erbj@stud.ntnu.no,simonhou@stud.ntnu.no"
LABEL stage=builder

WORKDIR /app

RUN apt-get update && apt-get install -y libproj-dev pkg-config

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY data/Losmasse/superficialdeposits_shape.zip data/Losmasse/superficialdeposits_shape.zip
COPY data/Fjord/fjordkatalogen_omrade.zip data/Fjord/fjordkatalogen_omrade.zip
COPY . .

RUN ls -la data/Losmasse

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /api ./cmd/api/main.go

FROM ubuntu:25.04

WORKDIR /root/

# Install runtime dependencies
RUN apt-get update && apt-get install -y libproj-dev unzip && touch .env

# Copy the built binary from builder stage
COPY --from=builder /api /api
COPY --from=builder /app/proxy.json proxy.json
COPY --from=builder /app/data/Losmasse data/Losmasse
COPY --from=builder /app/data/Fjord data/Fjord

RUN ls -la data/Losmasse

RUN ./data/Losmasse/prepare_data.sh ./data/Losmasse
RUN ./data/Fjord/prepare_data.sh ./data/Fjord

EXPOSE 8080

CMD [ "/api" ]
