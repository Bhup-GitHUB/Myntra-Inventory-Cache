FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN go build -o /out/api ./cmd/api
RUN go build -o /out/worker ./cmd/worker
RUN go build -o /out/seed ./cmd/seed

FROM alpine:3.20

WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=build /out/api /app/api
COPY --from=build /out/worker /app/worker
COPY --from=build /out/seed /app/seed
COPY migrations /app/migrations

EXPOSE 8080
CMD ["/app/api"]
