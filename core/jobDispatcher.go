package core

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type Job struct {
	ID     string
	Cmd    string
	User   string
	State  string
	cmdObj *exec.Cmd
}

func (j Job) ToString() string {
	return fmt.Sprintf("ID: %s, Cmd: %s, User: %s, State: %s", j.ID, j.Cmd, j.User, j.State)
}

const (
	Created  = "created"
	Running  = "running"
	Finished = "finished"
)

type JobDispatcher struct {
	// map of job id to job initialized as empty
	jobs map[string]Job
}

func (jd *JobDispatcher) Init() {
	jd.jobs = make(map[string]Job)
}

func (jd *JobDispatcher) ListJobs() []string {
	jobs := make([]string, len(jd.jobs))
	for _, job := range jd.jobs {
		jobs = append(jobs, job.ToString())
	}
	return jobs
}

func (jd *JobDispatcher) StopJob(jobId string) string {
	job := jd.jobs[jobId]
	// if job is not found, print a message and return
	if job.ID == "" {
		return "Job not found"
	}
	if job.State == Running {
		// Check if cmdObj is not nil and has a valid process
		if job.cmdObj != nil && job.cmdObj.Process != nil {
			// Attempt to kill the process
			err := job.cmdObj.Process.Kill()
			println("killing process")
			if err != nil {
				return "Failed to kill the process:"
			}
			job.State = Finished
		} else {
			return "No process to kill or command was not started"
		}
		return job.ToString()
	} else {
		return "Job is not running"
	}
}

func (jd *JobDispatcher) QueryJob(jobId string) string {
	job := jd.jobs[jobId]
	if job.ID == "" {
		return "Job not found"
	}
	return job.ToString()
}

func (jd *JobDispatcher) StartJob(job Job) string {
	jd.jobs[job.ID] = job
	job.State = Running

	cmdObj := exec.Command("sh", "-c", job.Cmd)
	job.cmdObj = cmdObj
	jd.jobs[job.ID] = job

	// 将io输入重定向到缓冲区
	var outBuf bytes.Buffer
	cmdObj.Stdout = &outBuf // 将io输入重定向到缓冲区

	// Start the command (non-blocking)
	err := cmdObj.Start()
	if err != nil {
		job.State = Finished
		jd.jobs[job.ID] = job
		return "Failed to start job:"
	}

	// Run the command in a goroutine
	err = cmdObj.Wait() // Wait for the command to finish
	if err != nil {
		return "Job finished with error:" + err.Error()
	} else {
		job.State = Finished
		jd.jobs[job.ID] = job
		return strings.TrimSpace(outBuf.String())
	}
}
