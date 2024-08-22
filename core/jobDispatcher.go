package core

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"os/exec"
	"strings"
	"sync"
)

type Job struct {
	ID     string
	Cmd    string
	User   string
	State  string
	cmdObj *exec.Cmd
}

type JobStatus struct {
	Job      *Job
	ExitCode int
	ErrorMsg string
}

func (j Job) ToString() string {
	return fmt.Sprintf("ID: %s, Cmd: %s, User: %s, State: %s", j.ID, j.Cmd, j.User, j.State)
}

func (js JobStatus) ToString() string {
	return fmt.Sprintf("Job: %s, ExitCode: %d, ErrorMsg: %s", js.Job.ToString(), js.ExitCode, js.ErrorMsg)
}

const (
	Created  = "created"
	Running  = "running"
	Finished = "finished"
)

type JobDispatcher struct {
	// map of job id to job initialized as empty
	jobs        map[string]*Job       // contain job info
	JobStatuses map[string]*JobStatus // contain job info and exit status
	lock        sync.RWMutex          // read write lock
}

func NewJobDispatcher() JobDispatcher {
	jd := JobDispatcher{}
	jd.Init()
	return jd
}

func (jd *JobDispatcher) Init() {
	jd.jobs = make(map[string]*Job)
	jd.JobStatuses = make(map[string]*JobStatus)
	// lru
}

func validateJobId(jobId string) error {
	if _, err := uuid.Parse(jobId); err != nil {
		return err
	}
	return nil
}

// list all jobs
func (jd *JobDispatcher) ListJobs() []JobStatus {
	jd.lock.RLock()
	defer jd.lock.RUnlock()
	jobs := make([]JobStatus, 0)
	for _, job := range jd.JobStatuses {
		jobs = append(jobs, *job)
	}
	return jobs
}

func (jd *JobDispatcher) StopJob(jobId string) string {
	// if job is not found, print a message and return
	if err := validateJobId(jobId); err != nil {
		return "Invalid job ID"
	}
	jd.lock.Lock()
	job := jd.jobs[jobId]
	jd.lock.Unlock()
	if job.State == Running {
		// Check if cmdObj is not nil and has a valid process
		if job.cmdObj != nil && job.cmdObj.Process != nil {
			// Attempt to kill the process
			jd.lock.Lock()
			err := job.cmdObj.Process.Kill()
			println("killing process")
			if err != nil {
				return "Failed to kill the process:"
			}
			job.State = Finished
			jobStatus := jd.JobStatuses[jobId]
			jobStatus.ExitCode = 1
			jobStatus.ErrorMsg = "signal: killed" // e.g. sleep 50 && pwd
			jd.lock.Unlock()
		} else {
			return "No process to kill or command was not started"
		}
		return job.ToString()
	} else {
		return "Job is not running"
	}
}

func (jd *JobDispatcher) QueryJob(jobId string) JobStatus {
	jd.lock.RLock()
	defer jd.lock.RUnlock()
	// if job is not found, print a message and return
	if err := validateJobId(jobId); err != nil {
		return JobStatus{Job: &Job{ID: jobId}, ExitCode: -1, ErrorMsg: "Invalid job ID"}
	}
	jobStatus := jd.JobStatuses[jobId]
	if jobStatus == nil {
		return JobStatus{Job: &Job{ID: jobId}, ExitCode: -1, ErrorMsg: "Job not found"}
	}
	return *jobStatus
}

func (jd *JobDispatcher) StartJob(job Job) string {
	jd.lock.Lock()
	job.State = Running
	cmdObj := exec.Command("sh", "-c", job.Cmd) // Create a new command object, prepare to run the command
	job.cmdObj = cmdObj
	jd.jobs[job.ID] = &job
	jobStatus := JobStatus{Job: &job, ExitCode: -1, ErrorMsg: ""}
	jd.JobStatuses[job.ID] = &jobStatus
	jd.lock.Unlock()
	// 将io输入重定向到缓冲区
	var outBuf bytes.Buffer
	cmdObj.Stdout = &outBuf // 将io输入重定向到缓冲区

	// Start the command (non-blocking)
	err := cmdObj.Start()
	if err != nil {
		jd.lock.Lock()
		job.State = Finished
		jobStatus.ExitCode = 1
		jobStatus.ErrorMsg = err.Error()
		jd.lock.Unlock()
		return "Failed to start job:"
	}
	// Run the command in a goroutine
	err = cmdObj.Wait() // Wait for the command to finish
	if err != nil {
		jd.lock.Lock()
		job.State = Finished
		jobStatus.ExitCode = 1
		jobStatus.ErrorMsg = err.Error() // sleep 50
		jd.lock.Unlock()
		return "Job finished with error:" + err.Error()
	} else {
		jd.lock.Lock()
		job.State = Finished
		jobStatus.ExitCode = 0
		jobStatus.ErrorMsg = ""
		jd.lock.Unlock()
		println(strings.TrimSpace(outBuf.String()))
		return strings.TrimSpace(outBuf.String())
	}
}
