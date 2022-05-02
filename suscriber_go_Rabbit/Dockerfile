FROM golang:1.17
WORKDIR /
COPY . .
RUN go mod download
CMD ["go","run","main.go"]