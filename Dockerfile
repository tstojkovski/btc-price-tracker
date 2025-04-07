FROM golang:1.23

WORKDIR /app

COPY . .

RUN go build -o server ./cmd/server/main.go

RUN chmod +x ./server

ENV STORE_SIZE=1000

EXPOSE 8082

CMD ["./server"]
