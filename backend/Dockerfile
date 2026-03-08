FROM golang:1.25 AS build

ARG SERVICE=api

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/service ./cmd/${SERVICE}

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=build /out/service /app/service

ENTRYPOINT ["/app/service"]
