FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -o /out/flashy-server ./cmd/server

FROM alpine:3.20
COPY --from=build /out/flashy-server /usr/local/bin/flashy-server
EXPOSE 8080
ENTRYPOINT ["flashy-server"]
