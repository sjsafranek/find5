FROM golang:latest
MAINTAINER Stefan Safranek, sjsafranek@gmail.com

COPY . /go/src/github.com/sjsafranek/find5/findapi
COPY config.json /etc/find5/config.json
RUN mkdir /var/log/find5

RUN go get github.com/sjsafranek/ligneous
RUN go get github.com/chzyer/readline
RUN go get github.com/garyburd/redigo/redis
RUN go get github.com/google/uuid
RUN go get github.com/gorilla/mux
RUN go get github.com/gorilla/sessions
RUN go get github.com/karlseguin/ccache
RUN go get github.com/lib/pq
RUN go get github.com/paulmach/orb/geojson
RUN go get github.com/sjsafranek/lemur
RUN go get github.com/sjsafranek/pool
RUN go get github.com/pkg/errors
RUN go get github.com/boltdb/bolt

RUN cd /go/src/github.com/sjsafranek/find5/findapi && go build -o /go/find *.go

CMD ./find -dbhost finddb -redishost redis -aihost findai -L /var/log/find5 -c /etc/find5/config.json
