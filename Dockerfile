FROM golang:1.24 AS builder

LABEL authors="bjornsen.erik@gmail.com,simonhou@stud.ntnu.no"
LABEL stage=builder

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
	libproj-dev \
	pkg-config \
	python3 \
	python3-pip \
	unzip \
	&& rm -rf /var/lib/apt/lists/*

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

COPY data/Losmasse/superficialdeposits_shape.zip data/Losmasse/superficialdeposits_shape.zip
COPY data/Losmasse/fix_invalid_values.py data/Losmasse/fix_invalid_values.py
COPY data/Losmasse/prepare_data.sh data/Losmasse/prepare_data.sh
COPY assets/forestry_road_legend.png assets/forestry_road_legend.png
COPY . .

RUN ls -la data/Losmasse

RUN ./data/Losmasse/prepare_data.sh ./data/Losmasse

RUN ls -la data/Losmasse

# RUN pip3 install dbf dbfread --break-system-packages
# RUN python3 ./data/Losmasse/fix_invalid_values.py ./data/Losmasse/LosmasseFlate_20240621.dbf

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /api ./cmd/api/main.go

FROM ubuntu:25.04

WORKDIR /root/

# Install runtime dependencies
RUN apt-get update && apt-get install -y libproj-dev unzip && touch .env

# Copy the built binary from builder stage
COPY --from=builder /api /api
COPY --from=builder /app/proxy.json proxy.json
COPY --from=builder /app/data/Losmasse data/Losmasse
COPY --from=builder /app/assets/forestry_road_legend.png assets/forestry_road_legend.png

RUN ls -la data/Losmasse

EXPOSE 8080

CMD [ "/api" ]
