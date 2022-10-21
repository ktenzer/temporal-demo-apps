FROM golang:1.18-alpine
RUN mkdir -p /app/bin
WORKDIR /app
ENV GOBIN=/app/bin
COPY . .

RUN go install ./...
CMD ["/app/bin/worker" ]
