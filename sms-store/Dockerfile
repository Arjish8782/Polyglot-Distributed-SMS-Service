FROM golang:1.26-alpine
WORKDIR /app

# 1. Copy ONLY the go modules first
COPY go.mod go.sum ./

# 2. Download the dependencies
RUN go mod download

# 3. Copy the rest of your source code
COPY . .

# 4. Build the application
RUN go build -o main .

# 5. Run the executable
CMD ["./main"]