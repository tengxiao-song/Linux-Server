package command

import (
	"main/core"

	"github.com/spf13/cobra"
)

func QueryCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "query",
		Short: "Query a job",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			jd := core.JobDispatcher{}
			res := jd.StopJob(args[0])
			println(res)
		},
	}
}
