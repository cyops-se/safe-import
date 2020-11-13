package rippers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/cyops-se/safe-import/si-outer/types"

	"github.com/PuerkitoBio/goquery"
)

type HttpContext struct {
	Context
	processed map[string]bool
	Url       *url.URL
}

func CreateHttpContext(job *types.Job) *HttpContext {
	context := &HttpContext{}
	context.Job = job
	context.processed = make(map[string]bool, 1)
	context.Url, _ = url.Parse(job.Repository.URL)
	return context
}

func (context *HttpContext) DownloadDirHttp() error {
	job := context.Job

	log.Printf("Downloading request URL %s\n", job.Repository.URL)
	job.LocalPath = path.Join(job.Repository.OuterPath, context.Url.Path)
	if strings.HasSuffix(job.LocalPath, "/") {
		job.LocalPath = path.Join(job.LocalPath, "index.html")
	}

	fmt.Println("OUTERPATH:", job.Repository.OuterPath)
	fmt.Println("LOCALPATH:", job.LocalPath)

	context.MkdirIfNotExists(job.LocalPath)

	context.RemoteTree = &Folder{Name: "/"}
	err := context.buildTreeFromURL(job.Repository.URL, context.RemoteTree)
	if job.ReportError(err, "DownloadDirHttp:context.buildTreeFromURL()") {
		return err
	}

	err = context.CheckDestinationPath(context.RemoteTree)
	if job.ReportError(err, "DownloadDirHttp:context.CheckDestinationPath()") {
		return err
	}

	fmt.Println("Downloading missing or modified files: ", context.Files)
	for f := context.Files.Front(); f != nil; f = f.Next() {
		select {
		case job.Command = <-job.Commands:
			err := fmt.Errorf("Job aborted by command: %d", job.Command)
			job.ReportError(err, "DownloadDirHttp:STOP COMMAND")
			return err
		default:
			file := f.Value.(*MissingFile)
			context.downloadFileHttp(file)
		}

		if job.Progress.Error != nil {
			return job.Progress.Error
		}
	}

	return nil
}

func (context *HttpContext) DownloadSingleFileHttp() error {
	job := context.Job
	prefix := "/safe-import/outer"

	log.Printf("Downloading request URL %s\n", job.Repository.URL)
	if len(strings.TrimSpace(job.LocalPath)) == 0 {
		job.LocalPath = fmt.Sprintf("%d", job.Repository.ID)
	}

	job.LocalPath = path.Join(prefix, job.LocalPath)

	context.MkdirIfNotExists(job.LocalPath)

	fmt.Println("Downloading requested file: ", job.Repository.URL)

	filename := context.Url.Path
	if filename == "" || filename == "/" {
		filename = "index.html"
	}

	context.downloadSingleFileHttp(job.Repository.URL, filename)

	return nil
}

func (context *HttpContext) buildTreeFromURL(urlstr string, folder *Folder) error {
	job := context.Job

	if job.Command > 0 {
		return nil
	}

	select {
	case job.Command = <-job.Commands:
		err := fmt.Errorf("Job aborted by command: %d", job.Command)
		job.ReportError(err, "buildTreeFromURL:STOP COMMAND")
		return err

	default:
		doc, err := goquery.NewDocument(urlstr)
		if job.ReportError(err, "buildTreeFromURL:goquery.NewDocument()") {
			return err
		}

		// fmt.Println("Ripping document:", urlstr)
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			href, ok := s.Attr("href")
			u, err := url.Parse(href)
			if err != nil || (u.Host != "" && u.Host != context.Url.Host) {
				return
			}

			href = u.RequestURI()

			if ok && !strings.Contains(href, "..") && !strings.HasPrefix(href, "/") && !strings.HasPrefix(href, "?") && !strings.HasPrefix(href, "http") {
				if strings.HasSuffix(href, "/") {
					if _, ok := context.processed[href]; !ok {
						if folder.Folders == nil {
							folder.Folders = make(map[string]*Folder, 1)
						}

						child := &Folder{Name: href}
						folder.Folders[child.Name] = child
						context.processed[href] = true
						if err := context.buildTreeFromURL(job.Repository.URL+href, child); err != nil {
							job.Progress.Error = err
							return
						}
					}
				} else {
					if _, ok := context.processed[href]; !ok {
						if folder.Files == nil {
							folder.Files = make(map[string]*File, 1)
						}

						context.processed[href] = true
						url := fmt.Sprintf("%s%s%s", context.Job.Repository.URL, folder.Name, href)
						if err := context.buildFileFromURL(url, folder); err != nil {
							job.Progress.Error = err
							return
						}
					}
				}
			}
		})

		if job.Progress.Error != nil {
			err = job.Progress.Error
		}

		return err
	}
}

func (context *HttpContext) buildFileFromURL(url string, folder *Folder) error {
	job := context.Job

	headResp, err := http.Head(url)
	if job.ReportError(err, "downloadFileHttp:http.Head()") {
		return err
	}
	defer headResp.Body.Close()

	if headResp.StatusCode != 200 {
		err := fmt.Errorf("Failed to retrieve HEAD for url: '%s', code: %d", url, headResp.StatusCode)
		job.ReportError(err, "downloadFileHttp:http.Head()")
		return err
	}

	filename := path.Base(url)
	folder.Files[filename] = &File{Name: filename}

	contentlength := headResp.Header.Get("Content-Length")
	if size, err := strconv.Atoi(contentlength); err == nil {
		folder.Files[filename].Size = int64(size)
	}

	return err
}

func (context *HttpContext) downloadFileHttp(f *MissingFile) error {
	job := context.Job
	fullURL := job.Repository.URL + f.Fullname

	localfile := path.Join(job.LocalPath, f.Fullname)
	context.MkdirIfNotExists(path.Dir(localfile))

	fmt.Println("Storing file at:", localfile)
	out, err := os.Create(localfile)
	if job.ReportError(err, "downloadFileHttp:os.Create()") {
		return err
	}
	defer out.Close()

	c := &http.Client{Timeout: -1}

	// Start a goroutine that checks local file size to report progress
	done := make(chan int64)
	job.Progress.CurrentPath = f.Fullname
	job.Progress.Current.Percent = 0.0
	job.Progress.Current.Total = int64(f.File.Size)
	go context.PrintDownloadPercent(done)

	resp, err := c.Get(fullURL)
	if job.ReportError(err, "downloadFileHttp:c.Get()") {
		done <- 1
		return err
	}

	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)
	out.Close()
	job.ReportError(err, "downloadFileHttp:io.Copy()")

	done <- n

	// fmt.Println("downloadFileHttp() DONE!")
	return err
}

func (context *HttpContext) downloadSingleFileHttp(url string, filename string) error {
	job := context.Job

	localfile := path.Join(job.LocalPath, filename)
	context.MkdirIfNotExists(path.Dir(localfile))

	fmt.Println("Storing file at:", localfile)
	out, err := os.Create(localfile)
	if job.ReportError(err, "downloadFileHttp:os.Create()") {
		return err
	}
	defer out.Close()

	c := &http.Client{Timeout: -1}

	// Start a goroutine that checks local file size to report progress
	done := make(chan int64)
	job.Progress.CurrentPath = filename
	job.Progress.Current.Percent = 0.0
	job.Progress.Current.Total = 0
	go context.PrintDownloadPercent(done)

	resp, err := c.Get(job.Repository.URL)
	if job.ReportError(err, "downloadFileHttp:c.Get()") {
		done <- 1
		return err
	}

	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)
	out.Close()
	job.ReportError(err, "downloadFileHttp:io.Copy()")

	done <- n

	// fmt.Println("downloadFileHttp() DONE!")
	return err
}
