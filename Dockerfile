# Stage 1: Build --------------------------

FROM golang:latest AS build

WORKDIR /app
ADD . /app

RUN go build -o proxyapp

# Stage 2: Application --------------------

FROM alpine

RUN apk update \
  && apk add ca-certificates \
  && rm -rf /var/cache/apk/*

WORKDIR /app
COPY --from=build /app/proxyapp /app/proxyapp

RUN chmod +x /app/proxyapp

EXPOSE 80 443

CMD [ "/app/proxyapp" ]
