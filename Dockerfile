FROM golang:1.27.1 AS builder

ARG GOOS=linux
ARG GOARCH=amd64
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-w -s" -a -installsuffix cgo -o main .

FROM alpine AS final
COPY --from=builder /app/main /main
ENTRYPOINT ["/main"]
