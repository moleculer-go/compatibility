FROM golang:1.22-alpine3.18
WORKDIR /app
COPY data-service.go /app/data-service.go
COPY go.mod /app/go.mod
COPY go.sum /app/go.sum
RUN go build /app/data-service.go
CMD [ "/app/data-service" ]