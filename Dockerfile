FROM golang:1.12-alpine
RUN echo "extra"
RUN apk update
RUN apk add git
RUN apk add tzdata
RUN cp /usr/share/zoneinfo/America/Denver /etc/localtime
RUN echo "extra"
ADD root /var/spool/cron/crontabs/root
RUN mkdir -p /go/src/postgres-scanner
ADD postgres-scanner.go  /go/src/postgres-scanner/postgres-scanner.go
ADD build.sh /build.sh
RUN chmod +x /build.sh
RUN /build.sh
CMD ["crond", "-f"]




