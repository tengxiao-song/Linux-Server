package main

import (
	"context"
	"google.golang.org/grpc"
	core "main/core"
	pb "main/proto"
	"net"
)

type server struct {
	pb.UnimplementedJobManagerServer
}

// global jobDispatcher
var jobDispatcher = core.NewJobDispatcher()

func (s *server) Start(ctx context.Context, in *pb.Job) (*pb.Job, error) {
	println("Received start request")
	// map the input to core.Job
	job := core.Job{
		ID:    in.ID,
		Cmd:   in.Cmd,
		User:  in.User,
		State: in.State,
	}
	go jobDispatcher.StartJob(job)
	return in, nil
}

func (s *server) Query(ctx context.Context, in *pb.JobID) (*pb.JobStatus, error) {
	println("Received query request")
	jobStatus := jobDispatcher.QueryJob(in.Id)
	pbJob := pb.Job{
		ID:    jobStatus.Job.ID,
		Cmd:   jobStatus.Job.Cmd,
		User:  jobStatus.Job.User,
		State: jobStatus.Job.State,
	}
	return &pb.JobStatus{
		Job:          &pbJob,
		ExitCode:     int32(jobStatus.ExitCode),
		ErrorMessage: jobStatus.ErrorMsg,
	}, nil
}

func (s *server) Stop(ctx context.Context, in *pb.JobID) (*pb.NilMessage, error) {
	println("Received stop request")
	jobDispatcher.StopJob(in.Id)
	return &pb.NilMessage{}, nil
}

func (s *server) List(ctx context.Context, in *pb.NilMessage) (*pb.JobStatusList, error) {
	println("Received list request")
	jobList := jobDispatcher.ListJobs()
	var pbJobStatusList []*pb.JobStatus
	println("jobList:", len(jobList))
	for _, jobStatus := range jobList {
		pbJob := pb.Job{
			ID:    jobStatus.Job.ID,
			Cmd:   jobStatus.Job.Cmd,
			User:  jobStatus.Job.User,
			State: jobStatus.Job.State,
		}
		pbJobStatus := pb.JobStatus{
			Job:          &pbJob,
			ExitCode:     int32(jobStatus.ExitCode),
			ErrorMessage: jobStatus.ErrorMsg,
		}
		pbJobStatusList = append(pbJobStatusList, &pbJobStatus)
	}
	return &pb.JobStatusList{JobStatusList: pbJobStatusList}, nil
}

func (s *server) StreamOutput(in *pb.JobID, stream pb.JobManager_StreamOutputServer) error {
	println("Received stream request")
	//jobDispatcher.StreamOutput(in.Id, stream)
	resultChan := make(chan string)
	go jobDispatcher.Output(in.Id, resultChan)
	for elem := range resultChan { // read data form core's channel
		// make string to []byte
		output := &pb.JobOutput{Output: []byte(elem)}
		err := stream.Send(output) // send data to client
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	listen, _ := net.Listen("tcp", ":8080")
	s := grpc.NewServer()
	pb.RegisterJobManagerServer(s, &server{})
	s.Serve(listen)
}
