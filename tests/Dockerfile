FROM mrmagooey/hpcaas-container-base:0.1.1

ENV GOLANG_VERSION 1.8.3

RUN apt-get update
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y wget git
RUN set -eux; \
	url="https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz"; \
	wget -nv -O go.tgz "$url"; \
	tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	export PATH="/usr/local/go/bin:$PATH"; \
	go version

ENV GOPATH /go
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR $GOPATH
