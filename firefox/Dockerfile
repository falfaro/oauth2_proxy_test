FROM jlesage/firefox
MAINTAINER Kimbro Staken

RUN apk update
RUN apk add ca-certificates bash git go libc-dev tcpdump
RUN go get golang.org/x/net/html
RUN go get github.com/PuerkitoBio/goquery
