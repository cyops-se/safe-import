package types

import "fmt"

type ItemProgress struct {
	Percent float64 `json:"percent"`
	Size    int64   `json:"size"`
	Total   int64   `json:"total"`
	Error   error   `json:"error"`
}

type JobProgress struct {
	ID          int          `json:"id"`
	CurrentPath string       `json:"path"`
	Current     ItemProgress `json:"current"`
	Total       ItemProgress `json:"total"`
	Error       error        `json:"error"`
}

type JobInfo struct {
	RepositoryID uint         `json:"repositoryid"`
	Progress     *JobProgress `json:"progress"`
}

const (
	IDLE = iota
	STARTING
	RUNNING
	PAUSED
	COMPLETED
	FAILED
)

type Job struct {
	RepositoryID uint
	LocalPath    string
	Progress     JobProgress
	Callback     func(*Job)
	Commands     chan int // 1 = stop
	Command      int
}

// TODO: Must find at better structure for this ...
func (job *Job) ReportError(err error, source string) bool {
	if err == nil {
		return false
	}

	fmt.Println("ERROR: ", source, err)
	job.Progress.Error = err
	if job.Callback != nil {
		job.Callback(job)
	}
	return true
}
