package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func main() {
	jd := JobDispatcher{}
	jd.run()
}

type JobDispatcher struct {
	// map of job id to job initialized as empty
	jobs map[string]Job
}

func (jd JobDispatcher) run() {
	jd.jobs = make(map[string]Job)
	var rootCmd = &cobra.Command{
		Use: "./main",
	}

	var cmd1 = &cobra.Command{
		Use:   "start [Args..]",
		Short: "run a linux command",
		Args:  cobra.MinimumNArgs(1), // 允许接受任意数量的参数
		Run: func(cmd *cobra.Command, args []string) {
			cmdStr := strings.Join(args, " ")
			job := Job{
				// create a uuid, unique id using timestamp
				ID:  fmt.Sprintf("%d", time.Now().Unix()),
				Cmd: cmdStr,
				// get the current user using the os package
				User:  os.Getenv("USER"),
				State: Created,
			}
			jd.jobs[job.ID] = job
			println(job.toString())
			go jd.runJob(job.ID)
			reader := bufio.NewReader(os.Stdin)
			for {
				fmt.Print("> ")
				userInput, _ := reader.ReadString('\n')
				userInput = strings.TrimSuffix(userInput, "\n")

				if userInput == "exit" {
					fmt.Println("exting...")
					break
				}

				// get the user command (the first word)
				jdCmd := strings.Split(userInput, " ")[0]
				jdArgds := strings.Split(userInput, " ")[1:]
				if jdCmd == "start" {
					job := Job{
						// create a uuid, unique id using timestamp
						ID:  fmt.Sprintf("%d", time.Now().Unix()),
						Cmd: strings.Join(jdArgds, " "),
						// get the current user using the os package
						User:  os.Getenv("USER"),
						State: Created,
					}
					jd.jobs[job.ID] = job
					println(job.toString())
					go jd.runJob(job.ID)
				} else if jdCmd == "stop" {
					go jd.stopJob(jdArgds[0])
				} else if jdCmd == "query" {
					go jd.queryJob(jdArgds[0])
				} else if jdCmd == "list" {
					go jd.listJobs()
				}
			}
		}}

	var cmd2 = &cobra.Command{
		Use:   "stop [job id]",
		Short: "stop a job",
		Args:  cobra.MatchAll(cobra.ExactArgs(1)),
		Run: func(cmd *cobra.Command, args []string) {
			jd.stopJob(args[0])
		},
	}

	var cmd3 = &cobra.Command{
		Use:   "query [job id]",
		Short: "query a job",
		Args:  cobra.MatchAll(cobra.ExactArgs(1)),
		Run: func(cmd *cobra.Command, args []string) {
			jd.queryJob(args[0])
		},
	}

	var cmd4 = &cobra.Command{
		Use:   "list",
		Short: "list all jobs",
		Args:  cobra.NoArgs,
		Run:   func(cmd *cobra.Command, args []string) { jd.listJobs() },
	}

	// 将命令添加到根命令
	rootCmd.AddCommand(cmd1)
	rootCmd.AddCommand(cmd2)
	rootCmd.AddCommand(cmd3)
	rootCmd.AddCommand(cmd4)

	rootCmd.Execute()
}

func (jd JobDispatcher) listJobs() {
	for _, job := range jd.jobs {
		println(job.toString())
	}
}

func (jd JobDispatcher) stopJob(jobId string) {
	job := jd.jobs[jobId]
	// if job is not found, print a message and return
	if job.ID == "" {
		fmt.Println("Job not found")
		return
	}
	if job.State == Running {
		// Check if cmdObj is not nil and has a valid process

		if job.cmdObj != nil && job.cmdObj.Process != nil {
			// Attempt to kill the process
			err := job.cmdObj.Process.Kill()
			println("killing process")
			if err != nil {
				fmt.Println("Failed to kill the process:", err)
				return
			}
			job.State = Finished
		} else {
			fmt.Println("No process to kill or command was not started")
		}
		println(job.toString())
	}
}

func (jd JobDispatcher) queryJob(jobId string) {
	job := jd.jobs[jobId]
	if job.ID == "" {
		fmt.Println("Job not found")
		return
	}
	println(job.toString())
}

func (jd *JobDispatcher) runJob(jobId string) {
	job := jd.jobs[jobId]
	job.State = Running

	cmdObj := exec.Command("sh", "-c", job.Cmd)
	job.cmdObj = cmdObj
	jd.jobs[jobId] = job

	// 将io输入重定向到缓冲区
	var outBuf bytes.Buffer
	cmdObj.Stdout = &outBuf // 将io输入重定向到缓冲区

	// Start the command (non-blocking)
	err := cmdObj.Start()
	if err != nil {
		fmt.Println("Failed to start job:", err)
		job.State = Finished
		jd.jobs[jobId] = job
		return
	}
	// Now that the command has started, Process should be non-nil
	if cmdObj.Process == nil {
		fmt.Println("Process is nil after starting command")
		job.State = Finished
		jd.jobs[jobId] = job
		return
	}

	// Run the command in a goroutine

	err = cmdObj.Wait() // Wait for the command to finish
	if err != nil {
		fmt.Println("Job finished with error:", err)
	} else {
		fmt.Println(strings.TrimSpace(outBuf.String()))
	}
	job.State = Finished
	jd.jobs[jobId] = job

}

type Job struct {
	ID     string
	Cmd    string
	User   string
	State  string
	cmdObj *exec.Cmd
}

func (j Job) toString() string {
	return fmt.Sprintf("ID: %s, Cmd: %s, User: %s, State: %s", j.ID, j.Cmd, j.User, j.State)
}

const (
	Created  = "created"
	Running  = "running"
	Finished = "finished"
)
