FROM golang:1.12-alpine
ARG GITHUB_ACCESS_TOKEN
ARG SERVICE_NAME

# Install some dependencies needed to build the project
RUN apk update \
    && apk add build-base bash ca-certificates git openssh-client postgresql \
    && git config --global url."https://${GITHUB_ACCESS_TOKEN}:@github.com/".insteadOf "https://github.com/"
ENV GO111MODULE=off
RUN go get -v gopkg.in/urfave/cli.v2 \
    && go get -v github.com/oxequa/realize \
    && go get -u -v github.com/calm/ssm-env \
    && go get -u -v github.com/onsi/ginkgo/ginkgo \
    && go get -u -v github.com/onsi/gomega \
    && go get -u -v github.com/modocache/gover \
    && go get -u -v github.com/mattn/goveralls \
    && go get -u -v github.com/pressly/goose/cmd/goose \
    && echo "Finished downloading dependencies"

WORKDIR /go/src/github.com/calm/${SERVICE_NAME}
ENV GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64
# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

# Copy entire repo contents to proper go formatted path
COPY . .

# Target the single executable to compile directly
RUN go build ./...
