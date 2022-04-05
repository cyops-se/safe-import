package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/cyops-se/safe-import/si-outer/rippers"
	"github.com/cyops-se/safe-import/si-outer/system"
	"github.com/cyops-se/safe-import/si-outer/types"
	"github.com/cyops-se/safe-import/usvc"
)

type JobsService struct {
	usvc.Usvc
	jobs    map[int]*types.Job
	repoSvc *usvc.UsvcStub
}

func (svc *JobsService) Initialize(broker *usvc.UsvcBroker) {
	svc.InitializeService(broker, 1, "si-outer", "jobs", "Handles execution and reporting of background workers called jobs (not Steve)")
	svc.RegisterMethod("allitems", svc.getAll)
	svc.RegisterMethod("deletebyid", svc.deleteJob)
	svc.RegisterMethod("requesturlwait", svc.requestUrlWait)
	svc.RegisterMethod("requestrepodownload", svc.requestRepoDownload)
	svc.repoSvc = usvc.CreateStub(broker, "repos", "si-inner", 1)

	svc.jobs = make(map[int]*types.Job, 1)
}

func (svc *JobsService) getAll(payload string) (interface{}, error) {
	var jobs []*types.Job
	for _, j := range svc.jobs {
		jobs = append(jobs, j)
	}
	// fmt.Println("ALL JOBS:", jobs)
	return jobs, nil
}

func (svc *JobsService) deleteJob(payload string) (interface{}, error) {
	request := &types.ByIdRequest{} // repo id
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		svc.LogGeneric("error", "Marshalling request from JSON failed: %#v", err)
		return nil, err
	}

	delete(svc.jobs, request.ID)

	var jobs []*types.Job
	for _, j := range svc.jobs {
		jobs = append(jobs, j)
	}
	// fmt.Println("ALL JOBS:", jobs)
	return jobs, nil
}

func (svc *JobsService) requestRepoDownload(payload string) (interface{}, error) {
	request := &types.ByIdRequest{} // repo id
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		svc.LogGeneric("error", "Marshalling request from JSON failed: %#v", err)
		return nil, err
	}

	msg := &types.ByIdRequest{ID: request.ID}
	response, err := svc.repoSvc.RequestMessage("byid", msg)
	if err != nil {
		svc.LogGeneric("error", "Request to  repository service failed: %#v", err)
		return nil, err
	}

	repo := &types.Repository{}
	if err := json.Unmarshal([]byte(response.Payload), &repo); err != nil {
		svc.LogGeneric("error", "Marshalling repository from JSON failed: %#v", err)
		return nil, err
	}

	job := &types.Job{}
	job.Commands = make(chan int)
	job.Progress.ID = request.ID
	job.Callback = func(job *types.Job) {
		// fmt.Printf("Publishing progress: %v\n", job.Progress)
		if job.Progress.Error == nil {
			if job.Command == 0 {
				svc.PublishData("event.jobs.progress", job)
			} else {
				svc.PublishData("event.jobs.stopping", job)
			}
		} else {
			svc.PublishData("event.jobs.failure", job)
		}
	}

	// request.Progress = &types.JobProgress{Current: &types.Progress{}, Total: &types.Progress{}}
	// fmt.Printf("jobs.runjob Un-marshalled JobRequest: %v\n", job)
	svc.jobs[request.ID] = job

	go func() {
		// Find 'inner' path and clear it
		if err := os.Remove(repo.InnerPath); err != nil && !os.IsNotExist(err) {
			svc.LogError("Failed to remove inner link, aborting:", err)
			return
		}

		repo.Available = false
		job.Progress.ErrorMessage = "DOWNLOADING"
		svc.PublishData("event.jobs.downloading", job)

		if strings.HasPrefix(repo.URL, "http") {
			context := rippers.CreateHttpContext(job, repo)
			if repo.Recursive {
				context.DownloadDirHttp()
			} else {
				context.DownloadSingleFileHttp()
			}
		} else if strings.HasPrefix(repo.URL, "smb") {
			context := rippers.CreateSmbContext(job, repo)
			context.DownloadDirectoryCifs()
		}

		if job.Progress.Error == nil {
			svc.PublishData("event.jobs.scanning", job)
			svc.LogInfo(fmt.Sprintf("Scanning: %s", repo.OuterPath))
			job.Progress.ErrorMessage = "SCANNING"
			if err, exitcode, out := system.Run(system.CMD_CLAMSCAN, repo.OuterPath); err != nil {
				log.Printf("CLAM exited with status code: %d, out: %s", exitcode, string(out))

				job.Progress.Error = err
				job.Progress.ErrorMessage = err.Error()

				text := string(out)
				if exitcode == 1 {
					// Infections are reported from ClamAV as two lines per found infection separated by \n (and possibly \r)
					// Lets report each infection individually
					info := strings.ReplaceAll(text, "\r", "") // First normalize the text by removing all \r
					parts := strings.Split(info, "\n")
					for i := 0; i < len(parts)-1; i += 2 {
						if runtime.GOOS == "windows" {
							job.Progress.CurrentPath = fmt.Sprintf("%s %s", strings.Split(parts[i], ":")[2], parts[i+1])
						} else {
							job.Progress.CurrentPath = parts[i]
						}
					}

					job.Progress.ErrorMessage = "INFECTED"
					svc.PublishData("event.jobs.infected", job)
					svc.LogError(fmt.Sprintf("Repository %d failed", repo.ID), fmt.Errorf("INFECTED data is not available at inner side, cause: %s", text))
					svc.LogInfection(fmt.Sprintf("Repository %d has INFECTED data", repo.ID), text)
				} else if exitcode == 2 {
					job.Progress.ErrorMessage = "FAILED"
					svc.PublishData("event.jobs.failed", job)
					svc.LogError(fmt.Sprintf("Repository %d failed", repo.ID), fmt.Errorf("Scan failed. Data is not available at inner side, cause: %s, %s", text, err.Error()))
				}

				job.Progress.Error = err
				return
			}
		}

		// Report job completed if there are no errors and link destination path to 'inner'
		if job.Progress.Error == nil {
			svc.PublishData("event.jobs.completed", job)
			folder := path.Dir(repo.InnerPath)
			os.MkdirAll(folder, os.ModeDir)
			// wd, _ := os.Getwd()
			log.Printf("linking %s with %s", repo.OuterPath, repo.InnerPath)
			if err := os.Symlink(repo.OuterPath, repo.InnerPath); err == nil {
				svc.LogInfo(fmt.Sprintf("Repository %d completed. Successful completion of job. Data now available at inner side", repo.ID))
			} else {
				svc.LogInfo(fmt.Sprintf("Repository %d completed. Failed to create inner link. Data is NOT available at inner side, cause: %s", repo.ID, err.Error()))
			}

		} else {
			svc.PublishData("event.jobs.failed", job)
			svc.LogError(fmt.Sprintf("Repository %d failed", repo.ID), fmt.Errorf("Data is not available at inner side, cause: %#v", job.Progress.Error))
			job.Progress.Error = err
		}

		if job.Command > 0 {
			svc.PublishData("event.jobs.stopped", job)
			svc.LogInfo(fmt.Sprintf("Repository %d stopped. Successful stop of job. Data is NOT available at inner side", repo.ID))
			job.Progress.Error = err
		}

		delete(svc.jobs, repo.ID)
	}()

	return &types.Response{Success: true, Message: fmt.Sprintf("Repository sync started, id: %d", repo.ID)}, nil
}

func (svc *JobsService) stopJob(payload string) (interface{}, error) {
	// // fmt.Printf("jobs.stopjob invoked with request: %s\n", string(payload))

	// request := &types.ByIdRequest{}
	// if err := json.Unmarshal([]byte(payload), &request); err != nil {
	// 	svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
	// 	return nil, err
	// }

	// if job, ok := svc.jobs[request.ID]; ok {
	// 	job.Commands <- 1 // 1 = stop, make a blocking call to get ack from job
	// 	delete(svc.jobs, request.ID)
	// }

	// return &types.Response{Success: true, Message: fmt.Sprintf("Job stopped (not really ... yet), id: %d", request.ID)}, nil

	return nil, fmt.Errorf("jobs.stopjob: not yet implemented")
}

func (svc *JobsService) requestUrlWait(payload string) (interface{}, error) {
	// fmt.Printf("jobs.requestUrlWait invoked with request: %s\n", string(payload))

	request := &types.WaitRequest{}
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
		return nil, err
	}

	u, _ := url.Parse(request.URL)
	// log.Println("URL Path:", u.Path)
	// TODO: check that the request URL matches the repo URL MatchURL field (regex)

	// Get the URI
	filename := strings.ReplaceAll(u.Path, "/", "_")
	localfile := path.Join("/safe-import/outer", u.Host, filename)
	if u.Path == "/" {
		localfile = path.Join(localfile, "index.html")
	}

	if err := os.MkdirAll(path.Dir(localfile), os.ModeDir|os.ModePerm); err != nil {
		svc.LogError(fmt.Sprintf("Failed to create directory: %s", localfile), err)
		return nil, err
	}

	// fmt.Println("Storing file at:", localfile)
	out, err := os.Create(localfile)
	if err != nil {
		svc.LogError(fmt.Sprintf("Failed to create file: %s", localfile), err)
		return nil, err
	}

	defer out.Close()

	c := &http.Client{Timeout: -1}

	decoded, err := base64.StdEncoding.DecodeString(request.Body)
	req, err := http.NewRequest(request.Method, request.URL, bytes.NewReader(decoded))
	for _, v := range request.Headers {
		req.Header.Add(v.Name, v.Value)
	}

	var resp *http.Response
	resp, err = c.Do(req)

	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, fmt.Errorf("Response object is nil")
	}

	// Copy response headers
	headers := make([]types.NameValue, 1)
	for n, v := range resp.Header {
		// // fmt.Println("name:", n, ", value:", v)
		headers = append(headers, types.NameValue{Name: n, Value: v[0]})
	}

	if resp.Body == nil {
		return nil, fmt.Errorf("Response body is nil")
	}

	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	out.Close()

	return &types.WaitResponse{Success: true, Filename: localfile, Error: err, Headers: headers}, nil
}

func (svc *JobsService) scan(localfile string) error {

	err, exitcode, out := system.Run(system.CMD_CLAMSCAN, localfile)
	if err != nil {
		// fmt.Printf("CLAM exited with status code: %d, out: %s", exitcode, string(out))
		text := string(out)
		if exitcode == 1 {
			// Infections are reported from ClamAV as two lines per found infection separated by \n (and possibly \r)
			// Lets report each infection individually
			info := strings.ReplaceAll(text, "\r", "") // First normalize the text by removing all \r
			parts := strings.Split(info, "\n")
			for i := 0; i < len(parts)-1; i += 2 {
				if runtime.GOOS == "windows" {
					// svc.LogInfection(strings.Split(parts[i], ":")[2], parts[i+1])
					// fmt.Println(strings.Split(parts[i], ":")[2], parts[i+1])
				} else {
					// svc.LogInfection(strings.Split(parts[i], ":")[1], parts[i+1])
					// fmt.Println(strings.Split(parts[i], ":")[2], parts[i+1])
				}
			}

			svc.LogGeneric("infection", "Infected file detected: INFECTED data is not available at inner side, cause: %s", text)
		} else if exitcode == 2 {
			svc.LogError("Anti-virus scan failed", fmt.Errorf("Scan failed. Data is not available at inner side, cause: %s", text))
		}
	}

	return err
}
