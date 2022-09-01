FROM golang:1.19 as build
WORKDIR /app
 ENV CGO_ENABLED=1 GOOS=linux GCCGO=gccgo CGO_LDFLAGS="-g -O2"
 COPY go.mod .
 COPY go.sum .

 RUN go mod download
COPY . .
RUN go build -a -o serverd cmd/serverd/main.go
# RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o serverd cmd/serverd/main.go

FROM gcr.io/distroless/base
COPY --from=build /app/serverd /
EXPOSE 8443

CMD ["/serverd"]