FROM golang:1.17
WORKDIR /
COPY . .
RUN go mod download
EXPOSE 3000
CMD ["go","run","main.go"]
