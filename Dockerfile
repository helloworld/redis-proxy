# DEVELOPMENT ONLY
FROM golang:1.9.2-alpine

ARG app_env
ENV APP_ENV $app_env

ENV CAPACITY=1000
ENV GLOBAL_EXPIRY=60000
ENV PORT=8080
ENV REDIS_ADDRESS=redis:6379
ENV MAX_CLIENTS=5

# Install git
RUN apk add --no-cache git mercurial

# Install dep (for dependencies)
RUN go get -u github.com/golang/dep/cmd/dep

# Install fresh (for reloading on code change)
RUN go get -u github.com/pilu/fresh

COPY . /go/src/github.com/helloworld/redis-proxy
WORKDIR /go/src/github.com/helloworld/redis-proxy

# DEVELOPMENT:
# $ dep ensure - Ensure availability of required libraries
# $ fresh - Automatically rebuild application in development
# CMD dep ensure && fresh

# PRODUCTION:
RUN dep ensure
RUN go build
CMD ./redis-proxy -capacity 1000 -global-expiry 6000 -port 8080 -max-clients 5 -redis-address http://redis-cache.app.render.com:10000

EXPOSE 8080

