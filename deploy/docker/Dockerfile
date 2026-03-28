FROM golang:1.23-alpine AS build
ARG APP_CMD=./cmd/gateway
WORKDIR /src
COPY . .
RUN go build -o /out/app ${APP_CMD}

FROM alpine:3.22
RUN addgroup -S app && adduser -S app -G app
USER app
WORKDIR /app
COPY --from=build /out/app /app/app
EXPOSE 8080
ENTRYPOINT ["/app/app"]
