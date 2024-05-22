FROM ubuntu:22.04

RUN apt-get update -y
RUN apt-get dist-upgrade -y

# create user "builder" with sudo privileges
ARG GID
ARG UID
ARG USER=builder
RUN groupadd --gid ${GID} $USER
RUN useradd --uid ${UID} --gid $USER --shell /bin/bash --home-dir /home/$USER --create-home $USER
RUN apt-get install -y sudo
RUN echo "builder ALL=(ALL) NOPASSWD: ALL" | tee -a /etc/sudoers

ENV TZ=Europe/Rome
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN apt-get install -y cmake gcc gcc-arm-none-eabi gcc-mingw-w64 git gzip \
                       libsodium-dev libusb-1.0-0 make pkg-config protobuf-compiler \
                       u-boot-tools vim wget

# install tamago-go
ENV TAMAGO_VERSION="1.22.3"
ENV TAMAGO_CHECKSUM="4caffcfbdacddcff466d798a2f962e68344f08aad7537ef3296c6a4467e7e1c5"
RUN wget -O tamago-go.tgz https://github.com/usbarmory/tamago-go/releases/download/tamago-go${TAMAGO_VERSION}/tamago-go${TAMAGO_VERSION}.linux-amd64.tar.gz
RUN echo "${TAMAGO_CHECKSUM} tamago-go.tgz" | sha256sum --strict --check -
RUN tar -C / -xzf tamago-go.tgz && rm tamago-go.tgz

# install minisign
ENV MINISIGN_VERSION="0.11"
ENV MINISIGN_CHECKSUM="74c2c78a1cd51a43a6c98f46a4eabefbc8668074ca9aa14115544276b663fc55"
RUN wget -O minisign.tgz https://github.com/jedisct1/minisign/archive/refs/tags/${MINISIGN_VERSION}.tar.gz
RUN echo "${MINISIGN_CHECKSUM} minisign.tgz" | sha256sum --strict --check -
RUN tar -xzf minisign.tgz
RUN cd minisign-${MINISIGN_VERSION}; mkdir build; cd build; cmake ..; make; make install
RUN rm -r minisign*

# install habtool
ENV HABTOOL_VERSION="v2023.04.12"
RUN su - $USER -c "/usr/local/tamago-go/bin/go install github.com/usbarmory/crucible/cmd/habtool@${HABTOOL_VERSION}"

ENV GOPATH "/home/${USER}/go"
ENV USBARMORY_GIT "/home/${USER}/usbarmory"
ENV TAMAGO "/usr/local/tamago-go/bin/go"
ENV PATH "${PATH}:${GOPATH}/bin:/usr/local/tamago-go/bin"

USER $USER
WORKDIR /build
