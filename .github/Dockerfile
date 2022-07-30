FROM goreleaser/goreleaser-cross:v1.18.3

RUN apt-get update \
 && apt-get install --no-install-recommends -y -q upx \
 && apt-get -y autoremove \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/* \
    /tmp/* \
    /var/tmp/* \
    rm -rf /usr/share/man/* \
    /usr/share/doc
