#!/bin/bash

which docker &> /dev/null

if (( $? )); then
  echo "Docker must be installed (and in your PATH) to use this build script. Exiting."
  exit 1
fi


USER_UID=$(id -u)
USERNAME=builder


if (( $USER_UID == 0 )); then
  USERNAME=root
fi


docker build -t ubuntu:focal-builder -f - . <<EOF
FROM ubuntu:focal

RUN ["/bin/bash", "-c", "if (( $USER_UID != 0 )); then \
  groupadd --gid $USER_UID $USERNAME \
  && useradd -s /bin/bash --uid $USER_UID --gid $USER_UID -m $USERNAME; fi"]

RUN apt update && apt install -y build-essential wget

ARG GOLANG_VERSION=1.21.1

RUN wget -O go.tgz https://golang.org/dl/go\${GOLANG_VERSION}.linux-amd64.tar.gz \
  && tar -C /usr/local -xzf go.tgz && rm go.tgz

ENV GOPATH /go
ENV PATH \$GOPATH/bin:/usr/local/go/bin:\$PATH

RUN mkdir -p "\$GOPATH/src" "\$GOPATH/bin" \
  && chmod -R 777 "\$GOPATH"
EOF


echo BUILDING...

docker run -it --rm \
  -v $(pwd):/workspace \
  -w /workspace \
  -u $USERNAME \
  ubuntu:focal-builder make bin/redirector

echo DONE BUILDING
