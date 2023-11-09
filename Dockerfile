FROM golang:1.21

WORKDIR /usr/src/workshop

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

EXPOSE 8080

CMD ["go", "run", "src/app.go"]
