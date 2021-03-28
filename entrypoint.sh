#!/bin/bash

if [ "$1" = 'server' ]; then

WATCH=${WATCH:-'.'}
fsnotify-exec -w $WATCH <<'EOF'
find /etc/fsnotify.d/ -name '*.sh' | xargs -n1 -I{} sh -c "echo {} && sh {}"
EOF

else
    exec "$@"
fi
