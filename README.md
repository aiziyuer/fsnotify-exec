fsnotify-exec
===

## 用法

``` bash
➜  fsnotify-exec git:(main) ✗ ./fsnotify-exec --help 
Usage:
  fsnotify-exec [flags]

Flags:
  -h, --help                   help for fsnotify-exec
      --ignore-glob strings    thd object which need be ignored(glob/wild).
      --ignore-regex strings   thd object which need be ignored(regex).
  -v, --version                version for fsnotify-exec
  -w, --watch strings          thd object which need be watched, eg: dir/file. (default [./])
```

## 可引用变量

- `NOTIFY_EVENT`标识是什么事件, 比如: `CHMOD`, `WRITE`等
- `NOTIFY_FILE`标识有变动的文件名

## 样例

``` bash
➜  fsnotify-exec git:(main) ✗ go build . && ./fsnotify-exec -w . <<'EOF'
echo "notify file: $NOTIFY_FILE"
EOF
```

## 真实场景

需要监听某个目录的文件变化, 对变化的文件需要重新计算hash并拷贝至新的位置(hash即新的文件名).

``` bash
➜  fsnotify-exec git:(main) ✗ go build . && ./fsnotify-exec -w . <<'EOF'
cat $NOTIFY_FILE | sha256sum | awk '{print $1}' | xargs -n1 -I{} \cp $NOTIFY_FILE ../{}
EOF
```

