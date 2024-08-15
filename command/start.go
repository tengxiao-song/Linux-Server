package command

import (
	"main/core"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func StartCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start a job",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// get the string of the command to run as a string
			cmdStr := strings.Join(args, " ")
			jd := core.JobDispatcher{}
			jd.Init()
			job := core.Job{
				ID: "1",
				// ID:    fmt.Sprintf("%d", time.Now().Unix()),
				Cmd:   cmdStr,
				User:  os.Getenv("USER"),
				State: core.Created,
			}
			res := jd.StartJob(job)
			if res != "" {
				println(res)
			}
		},
	}
}
