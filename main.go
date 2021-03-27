package ionotify_exec

import (
	"fmt"
	"github.com/aiziyuer/ionotify-exec/cmd"
	"github.com/aiziyuer/ionotify-exec/util"
)

func main() {
	util.SetupLogs(fmt.Sprintf("/var/log/%s/info.log", cmd.ProgramName))
	cmd.Execute()
}
