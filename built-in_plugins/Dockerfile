# FROM golang:1.19 as build

# ENV CGO_ENABLED=1 GOOS=linux GCCGO=gccgo CGO_LDFLAGS="-g -O2"

# WORKDIR /app
# RUN mkdir -p /app/plugins
# WORKDIR /app/plug1
#  RUN mkdir -p /app/plugins/plug1
#  COPY plug1/go.mod .
#  COPY plug1/go.sum .
#  RUN go mod download
#  COPY plug1/ .
#  RUN go build -buildmode=plugin -o /app/plugins/plug1/plug.so plug.go

# FROM alpine:3.9
# COPY --from=build /app/plugins/ /plugins/
# WORKDIR /usr/src


FROM alpine:3.9
COPY plug1/plug.so /plugins/plug1/