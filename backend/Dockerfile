FROM golang:1.22-alpine AS build
ARG GIT_COMMIT=dev
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN go mod tidy && CGO_ENABLED=0 go build -ldflags "-X main.GitCommit=${GIT_COMMIT}" -o /out/server .

FROM gcr.io/distroless/static
COPY --from=build /out/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
