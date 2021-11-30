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

func CreateHttpContext(job *types.Job, repo *types.Repository) *HttpContext {
	context := &HttpContext{}
	context.Job = job
	context.processed = make(map[string]bool, 1)
	context.Url, _ = url.Parse(repo.URL)
	context.Repository = repo
	return context
}

func (context *HttpContext) DownloadDirHttp() error {
	job := context.Job

	currentpath := context.Repository.OuterPath
	log.Printf("Downloading request URL %s\n", context.Repository.URL)
	if strings.HasSuffix(context.Repository.OuterPath, "/") {
		currentpath = path.Join(context.Repository.OuterPath, "index.html")
	}

	// fmt.Println("OUTER PATH:", context.Repository.OuterPath)
	// fmt.Println("CURRENT PATH:", currentpath)

	context.MkdirIfNotExists(currentpath)

	context.RemoteTree = &Folder{Name: "/"}
	err := context.buildTreeFromURL(context.Repository.URL, context.RemoteTree)
	if err != nil {
		// fmt.Println("DownloadDirHttp buildTreeFromURL ERROR:", err)
		return err
	}

	err = context.CheckDestinationPath(context.RemoteTree)
	if err != nil {
		// fmt.Println("DownloadDirHttp CheckDestinationPath ERROR:", err)
		return err
	}

	// fmt.Println("Downloading missing or modified files: ", context.Files)
	for f := context.Files.Front(); f != nil; f = f.Next() {
		select {
		case job.Command = <-job.Commands:
			err := fmt.Errorf("Job aborted by command: %d", job.Command)
			// fmt.Println("DownloadDirHttp aborted ERROR:", err)
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
	prefix := "/safe-import/outer"

	outerpath := path.Join("/safe-import/outer/", context.Repository.OuterPath)

	log.Printf("Downloading request URL %s\n", context.Repository.URL)
	if len(strings.TrimSpace(outerpath)) == 0 {
		outerpath = fmt.Sprintf("%d", context.Repository.ID)
	}

	outerpath = path.Join(prefix, outerpath)

	context.MkdirIfNotExists(outerpath)

	// fmt.Println("Downloading requested file: ", context.Repository.URL)

	filename := context.Url.Path
	if filename == "" || filename == "/" {
		filename = "index.html"
	}

	context.downloadSingleFileHttp(context.Repository.URL, filename)

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
		// fmt.Println("buildTreeFromURL aborted ERROR:", err)
		return err

	default:
		doc, err := goquery.NewDocument(urlstr)
		if err != nil {
			// fmt.Println("buildTreeFromURL goquery.NewDocument ERROR:", err)
			return err
		}

		// // fmt.Println("Ripping document:", urlstr)
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
						if err := context.buildTreeFromURL(context.Repository.URL+href, child); err != nil {
							job.Progress.Error = err
							job.Progress.ErrorMessage = err.Error()
							return
						}
					}
				} else {
					if _, ok := context.processed[href]; !ok {
						if folder.Files == nil {
							folder.Files = make(map[string]*File, 1)
						}

						context.processed[href] = true
						url := fmt.Sprintf("%s%s%s", context.Repository.URL, folder.Name, href)
						if err := context.buildFileFromURL(url, folder); err != nil {
							job.Progress.Error = err
							job.Progress.ErrorMessage = err.Error()
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
	headResp, err := http.Head(url)
	if err != nil {
		// fmt.Println("buildTreeFromURL http.Head ERROR:", err)
		return err
	}

	defer headResp.Body.Close()

	if headResp.StatusCode != 200 {
		err := fmt.Errorf("Failed to retrieve HEAD for url: '%s', code: %d", url, headResp.StatusCode)
		// fmt.Println("buildFileFromURL http.Head status code ERROR:", err)
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
	fullURL := context.Repository.URL + f.Fullname

	outerpath := path.Join("/safe-import/outer/", context.Repository.OuterPath)
	localfile := path.Join(outerpath, f.Fullname)
	context.MkdirIfNotExists(path.Dir(localfile))

	log.Printf("Storing file at: %s, (outerpath: %s)", localfile, outerpath)
	out, err := os.Create(localfile)
	if err != nil {
		// fmt.Println("downloadFileHttp os.Create ERROR:", err)
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
	if err != nil {
		done <- 1
		// fmt.Println("downloadFileHttp c.Get ERROR:", err)
		return err
	}

	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)
	out.Close()
	// fmt.Println("downloadFileHttp io.Copy ERROR:", err)

	done <- n

	// // fmt.Println("downloadFileHttp() DONE!")
	return err
}

func (context *HttpContext) downloadSingleFileHttp(url string, filename string) error {
	job := context.Job

	outerpath := path.Join("/safe-import/outer/", context.Repository.OuterPath)
	localfile := path.Join(outerpath, filename)
	context.MkdirIfNotExists(path.Dir(localfile))

	// fmt.Println("Storing file at:", localfile)
	out, err := os.Create(localfile)
	if err != nil {
		// fmt.Println("downloadSingleFileHttp os.Create ERROR:", err)
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

	resp, err := c.Get(context.Repository.URL)
	if err != nil {
		done <- 1
		// fmt.Println("downloadSingleFileHttp c.Get ERROR:", err)
		return err
	}

	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)
	out.Close()
	// fmt.Println("downloadSingleFileHttp io.Copy ERROR:", err)

	done <- n

	// // fmt.Println("downloadFileHttp() DONE!")
	return err
}
