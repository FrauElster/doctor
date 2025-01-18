FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o doctor

# Final stage
FROM alpine

RUN adduser -D -h /app appuser

WORKDIR /app
COPY --from=builder /app/doctor .

RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 8080
ENTRYPOINT ["./doctor"]