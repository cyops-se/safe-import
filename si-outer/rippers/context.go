package rippers

import (
	"container/list"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/cyops-se/safe-import/si-outer/types"
)

// CallbackFunction is used to report job progress like file, percentage, bytes done, bytes total
type CallbackFunction func(job *types.Job)

// File contain basic information about size and date
type File struct {
	Name string    `json:"name"`
	Date time.Time `json:"date"`
	Size int64     `json:"size"`
}

// Folder keeps a reference to files and folders contained in that folder
type Folder struct {
	Name    string             `json:"name"`
	Folders map[string]*Folder `json:"folders"`
	Files   map[string]*File   `json:"files"`
}

type MissingFile struct {
	Fullname string
	URIPath  string // part of URL after job.URL (filename or folder/filename)
	File     *File
}

// Context keep data specific for each ripping request
type Context struct {
	Job        *types.Job
	Files      *list.List // Files to download
	TotalCount int
	TotalSize  int64
	RemoteTree *Folder
}

func (context *Context) LoadTree(folder *Folder) {
	if folder == nil {
		folder = &Folder{Folders: make(map[string]*Folder, 1), Files: make(map[string]*File, 1)}
	}

	localdb := path.Join(context.Job.LocalPath, "check.json")
	if data, err := ioutil.ReadFile(localdb); err == nil {
		err = json.Unmarshal(data, folder)
	}

	return
}

func (context *Context) SaveTree(folder *Folder) error {
	j, _ := json.MarshalIndent(folder, "", "    ")
	localdb := path.Join(context.Job.LocalPath, "check.json")
	f, err := os.Create(localdb)
	if err == nil {
		defer f.Close()
		f.Write(j)
	}
	return err
}

func (context *Context) MkdirIfNotExists(folder string) error {
	_, err := ioutil.ReadDir(folder)
	if err != nil {
		err := os.MkdirAll(folder, os.ModeDir|os.ModePerm)
		if context.Job.ReportError(err, "context.MkdirIfNotExists()") {
			return err
		}
	}
	return err
}

func (folder *Folder) traverseFolder(context *Context, destination string, files *list.List, totalsize *int64) {
	job := context.Job
	basepath := job.LocalPath

	if job.Command != 0 {
		return
	}

	// fmt.Println("Traversing folder", path.Join(destination, folder.Name))

	select {
	case job.Command = <-job.Commands:
		err := fmt.Errorf("Job aborted by command: %d", job.Command)
		job.ReportError(err, "traverseFolder:STOP COMMAND")
		return
	default:
		// Check the folders in this folder
		for _, d := range folder.Folders {
			d.traverseFolder(context, path.Join(destination, folder.Name), files, totalsize)
		}

		// Now check files and add them as missing if they can't be found locally
		for _, f := range folder.Files {
			fullname := path.Join(destination, folder.Name, f.Name)
			fullpath := path.Join(basepath, fullname)
			// fmt.Println("Checking local file:", fullpath)
			if fi, err := os.Stat(fullpath); err != nil {
				// fmt.Println("Adding missing file:", fullname, f.Name)
				*totalsize += f.Size
				files.PushBack(&MissingFile{Fullname: fullname, File: f})
			} else if fi.Size() != f.Size {
				// fmt.Println("Adding file with different size:", fullname, fi.Size(), f.Name, f.Size)
				*totalsize += f.Size
				files.PushBack(&MissingFile{Fullname: fullname, File: f})
			}
		}
	}
}

func (context *Context) CheckDestinationPath(folder *Folder) error {
	context.Files = list.New()
	folder.traverseFolder(context, "/", context.Files, &context.TotalSize)
	context.TotalCount = context.Files.Len()
	context.Job.Progress.Total.Percent = 0
	context.Job.Progress.Total.Size = 0
	context.Job.Progress.Total.Total = context.TotalSize
	return nil
}

// ============================================================================
// Local helpers

func (f *File) equals(o *File) bool {
	return f.Size == o.Size && f.Date == o.Date && f.Name == o.Name
}

func (context *Context) PrintDownloadPercent(done chan int64) error {
	job := context.Job
	stop := false
	lastSize := int64(0)
	count := 0
	for {
		select {
		case <-done:
			stop = true
		default:
			if count > 250 { // Every 250 * 20 millisecs = every 5 sec
				filename := path.Join(job.LocalPath, job.Progress.CurrentPath)
				filename = strings.ReplaceAll(filename, "\\", "/")
				file, err := os.Open(filename)
				if job.ReportError(err, "printDownloadPercent:os.Open()") {
					return err
				}

				fi, err := file.Stat()
				file.Close()

				if job.ReportError(err, "printDownloadPercent:file.Stat()") {
					return err
				}

				size := fi.Size()
				job.Progress.Current.Size = size
				job.Progress.Total.Size += (size - lastSize)

				if size != lastSize {
					if size == 0 {
						size = 1
					}

					job.Progress.Current.Percent = float64(size) / float64(job.Progress.Current.Total) * 100
					job.Progress.Total.Percent = float64(job.Progress.Total.Size) / float64(job.Progress.Total.Total) * 100
					if job.Callback != nil {
						job.Callback(job)
					}
				}

				lastSize = size
				count = 0
			}

			time.Sleep(time.Millisecond * 20)
			count++
		}

		if stop {
			break
		}
	}
	return nil
}

func start() {
	/*
		tree = &Folder{Files: make(map[string]*File)}
		compareTree = loadTree()

		b = make(chan int)
		d = make(chan int)

		go func() {
			for {
				for active > 2 {

				}
				_ = <-b
				active++
				total++
			}
		}()

		go func() {
			for {
				_ = <-d
				done++
				active--
			}
		}()
	*/
}
