package tableaupublic

import (
	"errors"
	"fmt"
	"github.com/franela/goreq"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ErrWorkbookNotFound is returned when a request to
// download a workbook file doesn't return a workbook file.
var ErrWorkbookNotFound = errors.New("workbook not found")

// IsNotFound returns true or false whether the err
// is an ErrWorkbookNotFound error.
func IsNotFound(err error) bool {
	if err == ErrWorkbookNotFound {
		return true
	}

	return false
}

// Workbook contains information about a Tableau Public workbook.
type Workbook struct {
	RepoURL         string `json:"workbookRepoUrl"`
	Size            int64  `json:"size"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	ShowInProfile   bool   `json:"showInProfile"`
	AllowDataAccess bool   `json:"allowDataAccess"`
}

// DownloadWorkbookFile downloads the workbook to the directory and returns
// the full path to the downloaded file. The file name will
// be the repoURL plus a TWB or TWBX extension depending
// on the file type downloaded.
func DownloadWorkbookFile(repoURL string, directory string) (filename string, err error) {
	url := fmt.Sprintf("https://public.tableau.com/workbooks/%v?format=twb", repoURL)

	var res *http.Response
	res, err = http.Get(url)
	if err != nil {
		return
	}
	defer res.Body.Close()

	contentType := res.Header.Get("Content-Type")
	if !strings.EqualFold(contentType, "application/x-twb") {
		err = ErrWorkbookNotFound
		return
	}

	filename = filepath.Join(directory, repoURL)
	disposition := res.Header.Get("Content-Disposition")
	if strings.Contains(strings.ToLower(disposition), ".twbx") {
		filename += ".twbx"
	} else {
		filename += ".twb"
	}

	var out *os.File
	out, err = os.Create(filename)
	if err != nil {
		return
	}
	defer out.Close()

	_, err = io.Copy(out, res.Body)
	return
}

// AllWorkbooks gets a list of all workbooks for the username.
func AllWorkbooks(username string) ([]*Workbook, error) {
	const COUNT = 20
	var index = 0
	var all []*Workbook
	var res *goreq.Response
	var err error
	for {
		res, err = goreq.Request{
			Uri:    fmt.Sprintf("https://public.tableau.com/profile/api/%v/workbooks?no_cache=%v&index=%v&count=%v", username, time.Now().Unix(), index, COUNT),
			Accept: "application/json",
		}.Do()
		if err != nil {
			break
		}

		var workbooks []*Workbook
		err = res.Body.FromJsonTo(&workbooks)
		if err != nil {
			break
		}

		for _, w := range workbooks {
			all = append(all, w)
		}

		if len(workbooks) < COUNT {
			break
		}

		index += len(workbooks)
	}

	return all, err
}
