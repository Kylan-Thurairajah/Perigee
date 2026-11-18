FROM golang:1.22-alpine AS build
WORKDIR /src
COPY apps/api/go.mod apps/api/go.sum* ./
RUN if [ -f go.sum ]; then go mod download; fi
COPY apps/api/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /perigee-api .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /perigee-api /app/perigee-api
COPY data/fixtures /app/data/fixtures
EXPOSE 8080
CMD ["/app/perigee-api"]
