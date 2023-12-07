# Build stage
FROM golang:1.21.4-alpine3.18 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

#Run stage
FROM alpine:3.18.5
WORKDIR /app
COPY app.env .
COPY db/migration ./db/migration
COPY --from=builder /app/main .
EXPOSE 8080
CMD [ "/app/main" ]