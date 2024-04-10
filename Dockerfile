FROM golang:1.21.8

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN go build -o KeDuBack .

EXPOSE 8080

CMD ["./KeDuBack"]