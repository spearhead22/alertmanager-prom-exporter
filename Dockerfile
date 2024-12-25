FROM golang:1.19-alpine

WORKDIR /app
COPY . .

RUN go mod download
RUN go build -o alertmanager-exporter

FROM alpine:latest
WORKDIR /app
COPY --from=0 /app/alertmanager-exporter .

EXPOSE 8080
CMD ["./alertmanager-exporter"]