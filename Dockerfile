# DEVELOPMENT ONLY
FROM golang:1.9.2-alpine

ARG app_env
ENV APP_ENV $app_env

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
CMD ./redis-proxy --capacity $CAPACITY --global-expiry $GLOBAL_EXPIRY --port $PORT --max-clients $MAX_CLIENTS --redis-address $REDIS_ADDRESS

EXPOSE 8080

