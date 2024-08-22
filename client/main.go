package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	core "main/core"
	pb "main/proto"
	"os"
	"strings"
)

func startJob(c pb.JobManagerClient, job core.Job) {
	_, err := c.Start(context.Background(), &pb.Job{
		ID:    job.ID,
		Cmd:   job.Cmd,
		User:  job.User,
		State: job.State,
	})
	if err != nil {
		fmt.Println("Error starting job:", err)
		return
	}
	// print the response
	fmt.Println("Job started:", job.ToString())
}

func queryJob(c pb.JobManagerClient, jobID string) {
	jobStatus, err := c.Query(context.Background(), &pb.JobID{Id: jobID})
	//println(jobStatus.)
	if err != nil {
		fmt.Println("Error querying job:", err)
		return
	}
	// print the response
	fmt.Println("Job: ", jobStatus.Job)
	fmt.Println("Exit code:", jobStatus.ExitCode)
	fmt.Println("Error message:", jobStatus.ErrorMessage)
}

func stopJob(c pb.JobManagerClient, jobID string) {
	_, err := c.Stop(context.Background(), &pb.JobID{Id: jobID})
	if err != nil {
		fmt.Println("Error stopping job:", err)
		return
	}
	fmt.Println("ok")
}

func listJobs(c pb.JobManagerClient) {
	jobList, err := c.List(context.Background(), &pb.NilMessage{})
	if err != nil {
		fmt.Println("Error listing jobs:", err)
		return
	}
	// print the response
	for _, job := range jobList.JobStatusList {
		fmt.Println(job.Job)
		fmt.Println("Exit code:", job.ExitCode)
		fmt.Println("Error message:", job.ErrorMessage)
	}
}

func streamJobs(c pb.JobManagerClient, jobID string) {
	stream, err := c.StreamOutput(context.Background(), &pb.JobID{
		Id: jobID,
	})
	if err != nil {
		fmt.Println("Error streaming jobs:", err)
		return
	}
	go func() {
		for {
			jobStatus, err := stream.Recv() // read data from the stream from the server
			if err == io.EOF {
				break
			} else if err != nil {
				fmt.Println("Error receiving job status:", err)
				return
			}
			// print the response (current output)
			fmt.Println(jobStatus.String())
		}
	}()
}

func main() {
	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()
	//Use conn to call the server
	client := pb.NewJobManagerClient(conn)

	// get the string of the command to run from user input
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Enter a command (start, stop, query, list) and press Enter:")

	for {
		// Read input from the user
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		// Trim newline characters
		input = strings.TrimSpace(input)
		// Split the input into command and arguments
		parts := strings.Split(input, " ")
		if len(parts) < 1 {
			fmt.Println("Invalid input. Please enter a command.")
			continue
		}
		switch parts[0] {
		case "exit":
			os.Exit(0)
		case "start":
			if len(parts) < 2 {
				fmt.Println("Invalid input. Please enter a command to run.")
				continue
			}
			cmdStr := strings.Join(parts[1:], " ")
			job := core.Job{
				ID:    uuid.New().String(),
				Cmd:   cmdStr,
				User:  os.Getenv("USER"),
				State: core.Created,
			}
			startJob(client, job)
		case "query":
			if len(parts) < 2 {
				fmt.Println("Invalid input. Please enter a job ID.")
				continue
			}
			jobID := strings.TrimSpace(strings.Join(parts[1:], ""))
			queryJob(client, jobID)
		case "stop":
			if len(parts) < 2 {
				fmt.Println("Invalid input. Please enter a job ID.")
				continue
			}
			jobID := strings.TrimSpace(strings.Join(parts[1:], ""))
			stopJob(client, jobID)
		case "list":
			listJobs(client)
		case "stream":
			if len(parts) < 2 {
				fmt.Println("Invalid input. Please enter a job ID.")
				continue
			}
			jobID := strings.TrimSpace(strings.Join(parts[1:], ""))
			streamJobs(client, jobID)
		default:
			fmt.Println("Invalid command. Please enter a valid command.")
		}
	}
}
