// Copyright 2023 Huawei Cloud Computing Technologies Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package progress

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
)

type progressMsg float64

type ErrMsg struct{ Err error }

type progressWriter struct {
	total      int
	downloaded int
	file       *os.File
	reader     io.ReadCloser
	onProgress func(*tea.Program, float64)
}

func (pw *progressWriter) Start() error {
	// TeeReader calls pw.Write() each time a new response is received
	_, err := io.Copy(pw.file, io.TeeReader(pw.reader, pw))
	if err != nil {
		return errors.WithStack(err)
	}
	pw.file.Close()   // nolint:errcheck
	pw.reader.Close() // nolint:errcheck
	return nil
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	pw.downloaded += len(p)
	return len(p), nil
}

func getResponse(link string) (*http.Response, error) {
	var client *http.Client
	httpProxy := os.Getenv("HTTP_PROXY")
	if httpProxy != "" {
		proxyParsedURL, err := url.Parse(httpProxy)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		// new http client and set proxy
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyParsedURL),
			},
			Timeout: time.Hour,
		}
	} else {
		client = &http.Client{
			Timeout: time.Hour,
		}
	}

	resp, err := client.Get(link) // nolint:gosec
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("receiving status of %d for url: %s", resp.StatusCode, link)
	}
	return resp, nil
}

func NewDownloadProgram(prefix, pkgLink string, pkgPath string) error {
	resp, err := getResponse(pkgLink)
	if err != nil {
		return errors.WithStack(err)
	}
	//defer resp.Body.Close()
	if resp.ContentLength <= 0 {
		return errors.WithMessage(err, "can't parse content length, aborting download")
	}

	file, err := os.Create(pkgPath)
	if err != nil {
		return errors.WithMessage(err, "could not create file")
	}

	pw := &progressWriter{
		total:      int(resp.ContentLength),
		file:       file,
		reader:     resp.Body,
		onProgress: func(p *tea.Program, ratio float64) {},
	}

	// StartDownloadPkg the download
	return pw.Start()
}
