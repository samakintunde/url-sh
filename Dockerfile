ARG GO_VERSION=1.24

FROM node:alpine as web-builder

WORKDIR /web
COPY web .
RUN yarn
RUN yarn build

FROM golang:${GO_VERSION}-bookworm as builder

WORKDIR /usr/src/app
COPY go.mod ./
RUN go mod download && go mod verify
COPY . .
COPY --from=web-builder /web/dist /usr/src/app/web/dist
RUN GOARCH=amd64 GOOS=linux CGO_ENABLED=1 go build -v -o /run-app .


FROM debian:bookworm

RUN apt -y update
RUN apt -y install ca-certificates

COPY --from=builder /run-app /usr/local/bin/
CMD ["run-app"]
