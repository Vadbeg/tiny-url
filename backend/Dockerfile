FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o back-tiny-url main.go

EXPOSE 8080
# Run the compiled Go binary
CMD ["./back-tiny-url"]
