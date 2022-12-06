# Start by building the application.
FROM golang:1.19 as build

ARG environment=development
ARG build=undefined
ARG version=undefined

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags "-X 'main.environment=$environment' -X 'main.build=$build' -X 'main.version=$version'" -o /go/bin/app

# Now copy it into our base image.
FROM gcr.io/distroless/static-debian11
COPY --from=build /go/bin/app /
CMD ["/app"]
