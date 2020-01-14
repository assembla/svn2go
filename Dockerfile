FROM debian:stretch

VOLUME /app

RUN apt-get update && apt-get -y install git wget curl cmake make gcc g++ liblz4-1 pkg-config liblz4-dev \
    libutf8proc2 libutf8proc-dev libserf-1-1 libserf-dev libsqlite3-dev zlib1g zlib1g-dev && \
    cd /usr/src && \
    wget https://www-eu.apache.org/dist/subversion/subversion-1.13.0.tar.bz2 -q -O subversion.tar.bz2 && \
    tar xjf subversion.tar.bz2 && cd subversion-1.13.0 && \
    ./configure --without-berkeley-db && make -j4 && make install && ldconfig && \
		cd /usr/src && wget https://dl.google.com/go/go1.13.6.linux-amd64.tar.gz -q -O go.tar.gz && \
		tar xzf go.tar.gz -C /usr/local && ln -sf /usr/local/go/bin/go /usr/local/bin/

WORKDIR /app
CMD ["/bin/bash"]
