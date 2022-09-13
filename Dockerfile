# syntax=docker/dockerfile:1

FROM golang:1.19-alpine as build
WORKDIR /go/src/glim
RUN apk update && apk add --no-cache git gcc musl-dev
COPY go.mod .
COPY go.sum . 
RUN go mod download
RUN go mod verify
COPY . .
RUN go build -o /go/bin/glim

FROM alpine
RUN apk add ca-certificates
COPY --from=build /go/bin/glim /app/glim
RUN adduser --disabled-password glim
USER glim
EXPOSE 1323
EXPOSE 1636
CMD ["server", "start"]
ENTRYPOINT ["/app/glim"]