package rippers

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/cyops-se/safe-import/si-outer/types"
	"github.com/hirochachacha/go-smb2"
)

type SmbContext struct {
	Context
	processed map[string]bool
	Url       *url.URL
	rfs       *smb2.RemoteFileSystem
}

func CreateSmbContext(job *types.Job, repo *types.Repository) *SmbContext {
	context := &SmbContext{}
	context.Job = job
	context.processed = make(map[string]bool, 1)
	context.Url, _ = url.Parse(repo.URL)
	context.Repository = repo
	return context
}

func (context *SmbContext) DownloadDirectoryCifs() error {
	job := context.Job

	outerpath := context.Repository.OuterPath
	if len(strings.TrimSpace(outerpath)) == 0 {
		if u, _ := url.Parse(context.Repository.URL); u != nil {
			outerpath = u.RequestURI()
		} else {
			outerpath = "."
		}
	}

	context.MkdirIfNotExists(outerpath)

	log.Printf("Downloading files from '%s' to '%s'\n", context.Repository.URL, outerpath)

	conn, err := net.Dial("tcp", context.Url.Host+":445")
	if err != nil {
		// fmt.Println("DownloadDirectoryCifs net.Dial ERROR:", err)
		return err
	}

	defer conn.Close()

	di := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     context.Repository.Username,
			Password: context.Repository.Password,
		},
	}

	di.Negotiator.RequireMessageSigning = false
	c, err := di.Dial(conn)
	if err != nil {
		// fmt.Println("DownloadDirectoryCifs di.Dial ERROR:", err)
		return err
	}

	defer c.Logoff()

	share := strings.ReplaceAll(context.Url.Path, "/", `\`)
	share = fmt.Sprintf(`\\%s%s`, context.Url.Host, share)
	// fmt.Println("c.Mount(" + share + ")")
	rfs, err := c.Mount(share)
	if err != nil {
		// fmt.Println("DownloadDirectoryCifs c.Mount ERROR:", err)
		return err
	}

	defer rfs.Umount()
	context.rfs = rfs

	// DownloadDirCifs(p, ".", rfs, job)
	// save()
	// tree = &Folder{Files: make(map[string]*File)}

	context.RemoteTree = &Folder{Name: "/"}
	err = context.buildTreeFromURL("", context.RemoteTree)
	if err != nil {
		// fmt.Println("DownloadDirectoryCifs buildTreeFromURL ERROR:", err)
		return err
	}

	err = context.CheckDestinationPath(context.RemoteTree)
	if err != nil {
		// fmt.Println("DownloadDirectoryCifs CheckDestinationPath ERROR:", err)
		return err
	}

	// fmt.Println("Copying missing or modified files: ", context.Files)
	for f := context.Files.Front(); f != nil; f = f.Next() {
		select {
		case job.Command = <-job.Commands:
			err := fmt.Errorf("Job aborted by command: %d", job.Command)
			// fmt.Println("DownloadDirectoryCifs aborted ERROR:", err)
			return err
		default:
			file := f.Value.(*MissingFile)
			// fmt.Println("Processing file: ", file.Fullname)
			context.copyFile(file)
		}

		if job.Progress.Error != nil {
			return job.Progress.Error
		}
	}

	return nil
}

func (context *SmbContext) buildTreeFromURL(urlstr string, folder *Folder) error {
	job := context.Job
	urlstr = strings.ReplaceAll(urlstr, "/", `\`)

	if job.Command > 0 {
		return nil
	}

	select {
	case job.Command = <-job.Commands:
		err := fmt.Errorf("Job aborted by command: %d", job.Command)
		// fmt.Println("buildTreeFromURL aborted:", err)
		return err

	default:
		files, err := context.rfs.ReadDir(urlstr)
		if err == nil {
			for _, f := range files {
				href := f.Name()

				if f.IsDir() {
					if _, ok := context.processed[href]; !ok {
						if folder.Folders == nil {
							folder.Folders = make(map[string]*Folder, 1)
						}

						child := &Folder{Name: href}
						folder.Folders[child.Name] = child
						context.processed[href] = true
						// // fmt.Println("Folder", child.Name, "added to folder", folder.Name)
						if err := context.buildTreeFromURL(path.Join(urlstr, href), child); err != nil {
							job.Progress.Error = err
							job.Progress.ErrorMessage = err.Error()
							return err
						}
					}
				} else {
					if _, ok := context.processed[href]; !ok {
						if folder.Files == nil {
							folder.Files = make(map[string]*File, 1)
						}

						context.processed[href] = true
						// // fmt.Println("File", href, "added to folder", folder.Name)
						folder.Files[href] = &File{Name: href, Size: f.Size()}
					}
				}
			}
		}

		if job.Progress.Error != nil {
			err = job.Progress.Error
		}

		return err
	}
}

func (context *SmbContext) copyFile(file *MissingFile) error {
	job := context.Job
	target := path.Join(context.Repository.OuterPath, file.Fullname)
	// // fmt.Println("Copying file, from:", file.Fullname, ", to:", target)

	file.Fullname = strings.TrimLeft(file.Fullname, `/\`)
	file.Fullname = strings.ReplaceAll(file.Fullname, "/", `\`)
	d, err := context.rfs.Open(file.Fullname)
	if err == nil {
		defer d.Close()

		context.MkdirIfNotExists(path.Dir(target))
		out, err := os.Create(target)
		if err != nil {
			// fmt.Println("copyFile os.Create ERROR:", err)
			return err
		}

		defer out.Close()

		done := make(chan int64)
		job.Progress.CurrentPath = file.Fullname
		job.Progress.Current.Percent = 0.0
		job.Progress.Current.Total = int64(file.File.Size)
		go context.PrintDownloadPercent(done)

		n, err := io.Copy(out, d)
		out.Close()
		done <- n

		// fmt.Println("copyFile io.Copy ERROR:", err)

	} else {
		// fmt.Println("copyFile rfs.Open ERROR:", err)
	}

	return err
}

/*
// DownloadDirCifs downloads the directory
func DownloadDirCifs(base string, startDir string, rfs *smb2.RemoteFileSystem, job *types.JobRequest) {
	b := strings.ReplaceAll(base, `\\`, ``)
	b = strings.ReplaceAll(b, `\`, `/`)

	files, err := rfs.ReadDir(strings.ReplaceAll(startDir, `/`, `\`))
	if err == nil {
		for _, f := range files {
			if f.IsDir() {
				createDir(path.Join(b, startDir, f.Name()))
				tree.Files[path.Join(b, startDir, f.Name())] = &File{}
				compareTree.Files[path.Join(b, startDir, f.Name())] = tree.Files[path.Join(b, startDir, f.Name())]
				DownloadDirCifs(base, path.Join(startDir, f.Name()), rfs, job)
			} else {
				p := strings.ReplaceAll(path.Join(startDir, f.Name()), `/`, `\`)
				d, err := rfs.Open(p)
				if err == nil {
					defer d.Close()

					size := f.Size()
					tree.Files[path.Join(b, startDir, f.Name())] = &File{Name: f.Name(), Date: f.ModTime(), Bytes: size}

					if compareTree.Files[path.Join(b, startDir, f.Name())] == nil || !compareTree.Files[path.Join(b, startDir, f.Name())].equals(tree.Files[path.Join(b, startDir, f.Name())]) {
						compareTree.Files[path.Join(b, startDir, f.Name())] = tree.Files[path.Join(b, startDir, f.Name())]

						out, err := os.Create(path.Join(b, startDir, f.Name()))
						if err != nil {
							// fmt.Println(path.Join(b, startDir, f.Name()))
							panic(err)
						}
						defer out.Close()

						done := make(chan int64)
						job.Progress.CurrentPath = path.Join(b, startDir, f.Name())
						job.Progress.Current.Percent = 0.0
						job.Progress.Current.Total = int64(size)
						go printDownloadPercent(done, job)

						n, err := io.Copy(out, d)
						job.ReportError(err, "DownloadDirCifs:io.Copy()")
						// fmt.Println(n, "bytes copied to file")
						out.Close()
						done <- n
					}
				} else {
					// fmt.Println(p, "error")
				}
			}
		}
	} else {
		// fmt.Println(startDir, "error")
	}
}
*/
