FROM golang:1.22-alpine3.18
WORKDIR /app
COPY ./data-service.go ./app/data-service.go
COPY ./go.mod ./app/go.mod
COPY ./go.sum ./app/go.sum
RUN ls ./app/
# RUN go mod tidy
RUN go build data-service.go
CMD [ "/app/data-service" ]