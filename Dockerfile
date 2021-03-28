FROM golang AS build

WORKDIR /gomod
COPY . .
RUN CGO_ENABLED=0 go build -o /fsnotify-exec . 

FROM centos:7

COPY --from=build /fsnotify-exec /usr/bin/fsnotify-exec
WORKDIR /tmp

ENV WATCH "."
RUN mkdir -p /etc/fsnotify.d
COPY demo.sh /etc/fsnotify.d/demo.sh
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]
CMD ["server"]
