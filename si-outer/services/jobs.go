package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/cyops-se/safe-import/si-outer/system"
	"github.com/cyops-se/safe-import/si-outer/types"
	"github.com/cyops-se/safe-import/usvc"
)

type JobsService struct {
	usvc.Usvc
	jobs map[int]*types.Job
}

func (svc *JobsService) Initialize(broker *usvc.UsvcBroker) {
	svc.InitializeService(broker, 1, "si-outer", "jobs", "Handles execution and reporting of background workers called jobs (not Steve)")
	svc.RegisterMethod("allitems", svc.getAll)
	svc.RegisterMethod("runjob", svc.runJob)
	svc.RegisterMethod("stopjob", svc.stopJob)
	svc.RegisterMethod("requesturlwait", svc.requestUrlWait)

	svc.jobs = make(map[int]*types.Job, 1)
}

func (svc *JobsService) getAll(payload string) (interface{}, error) {
	return nil, fmt.Errorf("jobs.getall: not yet implemented")
}

func (svc *JobsService) runJob(payload string) (interface{}, error) {
	// request := &types.ByIdRequest{} // repo id
	// if err := json.Unmarshal([]byte(payload), &request); err != nil {
	// 	svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
	// 	return nil, err
	// }

	// repo := &types.Repository{}
	// if result := common.DB.First(&repo, request.ID); result.Error != nil {
	// 	svc.LogGeneric("error", "Error while accessing the database: %#v", result.Error)
	// 	return nil, result.Error
	// }

	// job := &types.Job{Repository: repo}
	// job.Commands = make(chan int)
	// job.Progress.ID = request.ID
	// job.Callback = func(job *types.Job) {
	// 	fmt.Printf("Publishing progress: %v\n", job.Progress)
	// 	if job.Progress.Error == nil {
	// 		if job.Command == 0 {
	// 			svc.PublishData("event.jobs.progress", job.Progress)
	// 		} else {
	// 			svc.PublishData("event.jobs.stopping", job.Progress)
	// 		}
	// 	} else {
	// 		svc.PublishData("event.jobs.failure", job.Progress)
	// 	}
	// }

	// // request.Progress = &types.JobProgress{Current: &types.Progress{}, Total: &types.Progress{}}
	// fmt.Printf("jobs.runjob Un-marshalled JobRequest: %v\n", job)
	// svc.jobs[request.ID] = job

	// go func() {
	// 	// Find 'inner' path and clear it
	// 	if err := os.Remove(job.Repository.InnerPath); err != nil && !os.IsNotExist(err) {
	// 		svc.LogError("Failed to remove inner link, aborting:", err)
	// 		return
	// 	}

	// 	job.Repository.Available = false
	// 	if result := common.DB.Save(&job.Repository); result.Error != nil {
	// 		svc.LogError(fmt.Sprintf("Job %d about to start. Failed to update repository data", job.Repository.ID), result.Error)
	// 	}

	// 	if strings.HasPrefix(job.Repository.URL, "http") {
	// 		context := rippers.CreateHttpContext(job)
	// 		if job.Repository.Recursive {
	// 			context.DownloadDirHttp()
	// 		} else {
	// 			context.DownloadSingleFileHttp()
	// 		}
	// 	} else if strings.HasPrefix(job.Repository.URL, "smb") {
	// 		context := rippers.CreateSmbContext(job)
	// 		context.DownloadDirectoryCifs()
	// 	}

	// 	if job.Progress.Error == nil {
	// 		svc.PublishData("event.jobs.scanning", job.Progress)
	// 		svc.LogInfo(fmt.Sprintf("Scanning: %s", job.LocalPath))
	// 		if err, exitcode, out := system.Run(system.CMD_CLAMSCAN, job.LocalPath); err != nil {
	// 			fmt.Printf("CLAM exited with status code: %d, out: %s", exitcode, string(out))

	// 			job.Progress.Error = err
	// 			text := string(out)
	// 			if exitcode == 1 {
	// 				// Infections are reported from ClamAV as two lines per found infection separated by \n (and possibly \r)
	// 				// Lets report each infection individually
	// 				info := strings.ReplaceAll(text, "\r", "") // First normalize the text by removing all \r
	// 				parts := strings.Split(info, "\n")
	// 				for i := 0; i < len(parts)-1; i += 2 {
	// 					if runtime.GOOS == "windows" {
	// 						// svc.LogInfection(strings.Split(parts[i], ":")[2], parts[i+1])
	// 						fmt.Println(strings.Split(parts[i], ":")[2], parts[i+1])
	// 					} else {
	// 						// svc.LogInfection(strings.Split(parts[i], ":")[1], parts[i+1])
	// 						fmt.Println(strings.Split(parts[i], ":")[2], parts[i+1])
	// 					}
	// 				}

	// 				svc.PublishData("event.jobs.infected", job.Progress)
	// 				svc.LogError(fmt.Sprintf("Job %d failed", job.Repository.ID), fmt.Errorf("INFECTED data is not available at inner side, cause: %s", text))
	// 			} else if exitcode == 2 {
	// 				svc.PublishData("event.jobs.failed", job.Progress)
	// 				svc.LogError(fmt.Sprintf("Job %d failed", job.Repository.ID), fmt.Errorf("Scan failed. Data is not available at inner side, cause: %s", text))
	// 			}
	// 			return
	// 		}
	// 	}

	// 	// Report job completed if there are no errors and link destination path to 'inner'
	// 	if job.Progress.Error == nil {
	// 		svc.PublishData("event.jobs.completed", job.Progress)
	// 		folder := path.Dir(job.Repository.InnerPath)
	// 		os.MkdirAll(folder, os.ModeDir)
	// 		wd, _ := os.Getwd()
	// 		if err := os.Symlink(filepath.VolumeName(wd)+job.LocalPath, job.Repository.InnerPath); err == nil {
	// 			job.Repository.LastSuccess = time.Now().UTC()
	// 			job.Repository.Available = true
	// 			if result := common.DB.Save(&job.Repository); result.Error != nil {
	// 				svc.LogError(fmt.Sprintf("Job %d completed. Failed to update repository data", job.Repository.ID), result.Error)
	// 			} else {
	// 				svc.LogInfo(fmt.Sprintf("Job %d completed. Successful completion of job. Data now available at inner side", job.Repository.ID))
	// 			}
	// 		} else {
	// 			svc.LogInfo(fmt.Sprintf("Job %d completed. Failed to create inner link. Data is NOT available at inner side, cause: %#v", job.Repository.ID, err))
	// 		}

	// 	} else {
	// 		svc.PublishData("event.jobs.failed", job.Progress)
	// 		svc.LogError(fmt.Sprintf("Job %d failed", job.Repository.ID), fmt.Errorf("Data is not available at inner side, cause: %#v", job.Progress.Error))
	// 	}

	// 	if job.Command > 0 {
	// 		svc.PublishData("event.jobs.stopped", job.Progress)
	// 		svc.LogInfo(fmt.Sprintf("Job %d stopped. Successful stop of job. Data is NOT available at inner side", job.Repository.ID))
	// 	}
	// }()

	// return &types.Response{Success: true, Message: fmt.Sprintf("Job started, id: %d", job.Repository.ID)}, nil

	return nil, fmt.Errorf("jobs.runjob: not yet implemented")
}

func (svc *JobsService) stopJob(payload string) (interface{}, error) {
	// fmt.Printf("jobs.stopjob invoked with request: %s\n", string(payload))

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
	fmt.Printf("jobs.requestUrlWait invoked with request: %s\n", string(payload))

	request := &types.WaitRequest{}
	if err := json.Unmarshal([]byte(payload), &request); err != nil {
		svc.LogGeneric("error", "Marshalling request to JSON failed: %#v", err)
		return nil, err
	}

	u, _ := url.Parse(request.URL)
	fmt.Println("URL Path:", u.Path)
	// TODO: check that the request URL matches the repo URL MatchURL field (regex)

	// Get the URI
	localfile := path.Join("/safe-import/outer", u.Path)
	if u.Path == "/" {
		localfile = path.Join(localfile, "index.html")
	}

	if err := os.MkdirAll(path.Dir(localfile), os.ModeDir|os.ModePerm); err != nil {
		svc.LogError(fmt.Sprintf("Failed to create directory: %s", localfile), err)
		return nil, err
	}

	fmt.Println("Storing file at:", localfile)
	out, err := os.Create(localfile)
	if err != nil {
		svc.LogError(fmt.Sprintf("Failed to create file: %s", localfile), err)
		return nil, err
	}
	defer out.Close()

	c := &http.Client{Timeout: -1}
	resp, err := c.Get(request.URL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	out.Close()

	err = svc.scan(localfile)

	return &types.WaitResponse{Success: true, Filename: localfile, Error: err}, nil
}

func (svc *JobsService) scan(localfile string) error {

	err, exitcode, out := system.Run(system.CMD_CLAMSCAN, localfile)
	if err != nil {
		fmt.Printf("CLAM exited with status code: %d, out: %s", exitcode, string(out))
		text := string(out)
		if exitcode == 1 {
			// Infections are reported from ClamAV as two lines per found infection separated by \n (and possibly \r)
			// Lets report each infection individually
			info := strings.ReplaceAll(text, "\r", "") // First normalize the text by removing all \r
			parts := strings.Split(info, "\n")
			for i := 0; i < len(parts)-1; i += 2 {
				if runtime.GOOS == "windows" {
					// svc.LogInfection(strings.Split(parts[i], ":")[2], parts[i+1])
					fmt.Println(strings.Split(parts[i], ":")[2], parts[i+1])
				} else {
					// svc.LogInfection(strings.Split(parts[i], ":")[1], parts[i+1])
					fmt.Println(strings.Split(parts[i], ":")[2], parts[i+1])
				}
			}

			svc.LogError("Infected file detected", fmt.Errorf("INFECTED data is not available at inner side, cause: %s", text))
		} else if exitcode == 2 {
			svc.LogError("Anti-virus scan failed", fmt.Errorf("Scan failed. Data is not available at inner side, cause: %s", text))
		}
	}

	return err
}
