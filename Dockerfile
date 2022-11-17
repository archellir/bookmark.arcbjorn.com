## Build
FROM golang:1.16-buster AS build_stage

WORKDIR /
COPY . .
RUN go build -o /app/build cmd/main.go

## Deploy
FROM gcr.io/distroless/base-debian10 as deploy_stage

WORKDIR /

COPY --from=build_stage /app/build /app/build
COPY .env .
COPY internal/db/migrations ./db/migrations

EXPOSE 8080
ENTRYPOINT [ "/app/build" ]
