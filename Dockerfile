FROM golang:1.23

WORKDIR /app

COPY . .

RUN go build -o server ./cmd/server/main.go

RUN chmod +x ./server

# Default to memory store
ENV STORE_TYPE=mongo
ENV STORE_SIZE=1000

# MongoDB settings (used when STORE_TYPE=mongo)
ENV MONGO_URI=mongodb://host.docker.internal:27017
ENV MONGO_DATABASE=btc_price_tracker
ENV MONGO_COLLECTION=price_updates
# 86400 = 24 * 60 * 60
ENV MONGO_TTL=86400

ENV PRICE_PROVIDER=BINANCE

EXPOSE 8082

CMD ["./server"]
