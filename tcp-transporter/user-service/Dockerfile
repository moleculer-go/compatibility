FROM golang:1.22-alpine3.18
WORKDIR /app
COPY ./*.go ./
COPY ./*.mod ./
COPY ./*.sum ./
# RUN go mod tidy
RUN go build user-service.go
CMD [ "/app/user-service" ]