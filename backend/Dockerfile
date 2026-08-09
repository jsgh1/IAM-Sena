FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
# Generate/complete go.sum inside the builder before compiling.
# This makes a clean ZIP build correctly even if go.sum is not present yet.
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/iam-api ./cmd/api

FROM alpine:3.21
RUN adduser -D -H appuser
USER appuser
COPY --from=build /out/iam-api /usr/local/bin/iam-api
EXPOSE 8080
ENTRYPOINT ["iam-api"]
