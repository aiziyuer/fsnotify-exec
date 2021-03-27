package main

import (
	"fmt"

	"github.com/aiziyuer/fsnotify-exec/cmd"
	"github.com/aiziyuer/fsnotify-exec/util"
)

func main() {
	util.SetupLogs(fmt.Sprintf("/var/log/%s/info.log", cmd.ProgramName))
	cmd.Execute()
}
