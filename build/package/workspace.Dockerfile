FROM golang:1.23-alpine AS build

WORKDIR /file-service

RUN apk add -U --no-cache ca-certificates

COPY ./file-service/go.mod .
COPY ./file-service/go.sum .

RUN go mod download

COPY . /

RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /service ./cmd/main.go

FROM scratch

WORKDIR /

COPY --from=build /service /

EXPOSE 8080

ENTRYPOINT [ "/service", "--env", ""]
