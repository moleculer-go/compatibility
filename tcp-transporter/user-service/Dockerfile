FROM golang:1.22-alpine3.18
WORKDIR /app
COPY user-service.go /app/user-service.go
COPY go.mod /app/go.mod
COPY go.sum /app/go.sum
RUN go mod tidy
RUN go build /app/user-service.go
CMD [ "/app/user-service" ]