FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /out/atm-web ./cmd/atm-web

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build /out/atm-web /app/atm-web
COPY data/cuentas.seed.json /app/data/cuentas.seed.json
EXPOSE 8080
ENTRYPOINT ["/app/atm-web", "-addr", ":8080"]
