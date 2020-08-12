FROM golang:latest as builder

LABEL maintainer="Cristiano Colangelo <cristiano.colangelo@bluewin.ch>"

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

######## Start a new stage from scratch #######
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080

# Command to run the executable
ENTRYPOINT ["./main"]