FROM golang:1.12

RUN DEBIAN_FRONTEND=noninteractive apt-get update && apt-get install -y \
  rsync \
  curl \
  git \
  mercurial \
  rsync \
  sudo

RUN curl -fL https://github.com/bitrise-io/bitrise/releases/download/1.31.0/bitrise-$(uname -s)-$(uname -m) > /usr/local/bin/bitrise
RUN chmod +x /usr/local/bin/bitrise
RUN bitrise setup

RUN mkdir -p /go/src/github.com/bitrise-io/bitrise-add-new-project
ADD . /go/src/github.com/bitrise-io/bitrise-add-new-project
WORKDIR /go/src/github.com/bitrise-io/bitrise-add-new-project

ENTRYPOINT bitrise run test