FROM golang:1.12
ENV workdir /build
WORKDIR $workdir
COPY . .

RUN go install -v .

VOLUME ["/data"]
WORKDIR /data
CMD ["wikifeedia"]

