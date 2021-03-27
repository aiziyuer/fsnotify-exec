FROM golang AS build

WORKDIR /gomod
COPY . .
RUN CGO_ENABLED=0 go build -o /fsnotify-exec . 

FROM centos:7

COPY --from=build /fsnotify-exec /usr/bin/fsnotify-exec
WORKDIR /tmp
CMD ["fsnotify-exec", "-w", ".", "echo $NOTIFY_FILE"]
