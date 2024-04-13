ARG GO_VERSION=1

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
RUN go build -v -o /run-app .


FROM debian:bookworm

COPY --from=builder /run-app /usr/local/bin/
CMD ["run-app"]
