package command

import (
	"main/core"

	"github.com/spf13/cobra"
)

func ListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all jobs",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			jp := core.JobDispatcher{}
			res := jp.ListJobs()
			for _, job := range res {
				println(job)
			}
		},
	}
}
