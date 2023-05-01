# Build Container
FROM golang:1.20-alpine AS build

RUN apk --update --no-cache add make

WORKDIR /app
ADD . /app

RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -o go-app cmd/main.go

# Run Image
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=build /app/go-app /go-app
USER 65532:65532

ENTRYPOINT ["/go-app"]
