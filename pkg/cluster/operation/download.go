// Copyright 2023 Huawei Cloud Computing Technologies Co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package operation

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/openGemini/gemix/util"
)

type DownloadOptions struct {
	Version string
	Os      string
	Arch    string
}

type Downloader interface {
	Run() error
	downloadFile() error
	decompressFile() error
}

type GeminiDownloader struct {
	website     string
	version     string
	typ         string
	Url         string
	destination string
	timeout     time.Duration
	fileName    string
}

func NewGeminiDownloader(ops DownloadOptions) Downloader {
	return &GeminiDownloader{
		website:     util.DownloadWeb,
		version:     ops.Version,
		typ:         "-" + ops.Os + "-" + ops.Arch + util.DownloadPkgSuffix,
		destination: util.DownloadDst,
		timeout:     util.DownloadTimeout,
	}
}

func (d *GeminiDownloader) spliceUrl() error {
	if d.website == "" {
		d.website = util.DownloadWeb
	}

	if d.version == "" {
		latestVer, err := util.GetLatestVerFromCurl()
		if err != nil {
			return err
		} else {
			d.version = latestVer
		}
	}

	d.Url = d.website + "/" + d.version + "/" + util.DownloadFillChar + d.version[1:] + d.typ
	return nil
}

func (d *GeminiDownloader) Run() error {
	if _, err := os.Stat(util.DownloadDst); os.IsNotExist(err) {
		errDir := os.MkdirAll(util.DownloadDst, 0750)
		if errDir != nil {
			return errDir
		}
	}

	if err := d.spliceUrl(); err != nil {
		return err
	}
	isExisted, err := d.isExistedFile()
	if err != nil {
		return err
	}
	if !isExisted { // check whether need to download the files
		if err := d.downloadFile(); err != nil {
			return err
		}
	}

	if err := d.decompressFile(); err != nil {
		return err
	}

	return nil
}

func (d *GeminiDownloader) isExistedFile() (bool, error) {
	dir := filepath.Join(d.destination, d.version)
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return false, err
	} else if len(fs) != 1 {
		return false, fmt.Errorf("more than one offline installation package file at %s", dir)
	}
	d.fileName = filepath.Join(dir, fs[0].Name())
	return true, nil
}

func (d *GeminiDownloader) downloadFile() error {
	dir := filepath.Join(d.destination, d.version)
	fmt.Printf("start downloading file from %s to %s\n", d.Url, dir)

	var client *http.Client
	// get HTTP_PROXY and parse
	httpProxy := os.Getenv("HTTP_PROXY")
	if httpProxy != "" {
		fmt.Printf("use HTTP_PROXY: %s\n", httpProxy)
		proxyParsedURL, err := url.Parse(httpProxy)
		if err != nil {
			fmt.Printf("parse httpProxy failed! %v\n", err)
			return err
		}
		// new http client and set proxy
		client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyParsedURL),
			},
			Timeout: d.timeout,
		}
	} else {
		client = &http.Client{
			Timeout: d.timeout,
		}
	}

	// new GET request
	req, err := http.NewRequest(http.MethodGet, d.Url, nil)
	if err != nil {
		fmt.Printf("new GET request failed! %v\n", err)
		return err
	}

	// send request with ctx
	ctx := req.Context()
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		fmt.Printf("send GET request failed! %v\n", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("GET request failed, status code: %d", resp.StatusCode)
	}

	// create local file
	d.CleanFile(dir)
	if err = os.Mkdir(dir, 0750); err != nil {
		return err
	}
	fmt.Printf("mkdir: %s\n", dir)
	idx := strings.LastIndex(d.Url, "/")
	dst := filepath.Join(dir, d.Url[idx+1:])
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	fmt.Printf("create file: %s\n", dst)
	d.fileName = dst
	defer out.Close()

	// write response to file
	_, err = io.Copy(out, resp.Body)
	fmt.Printf("finish downloading file from %s to %s\n", d.Url, dir)
	return err
}

func (d *GeminiDownloader) decompressFile() error {
	targetPath := filepath.Join(d.destination, d.version)
	fmt.Printf("start decompressing %s to %s\n", d.fileName, targetPath)

	file, err := os.Open(d.fileName)
	if err != nil {
		fmt.Println("Error opening source file:", err)
		return err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		fmt.Println("Error creating gzip reader:", err)
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println("Error reading tar header:", err)
			return err
		}

		targetFile := filepath.Join(targetPath, header.Name)

		if header.FileInfo().IsDir() {
			continue
		}

		targetDir := filepath.Dir(targetFile)

		if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
			return err
		}
		file, err := os.Create(targetFile)
		if err != nil {
			fmt.Println("Error creating target file:", err)
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, tarReader)
		if err != nil {
			fmt.Println("Error extracting file:", err)
			return err
		}
	}

	fmt.Printf("finish decompressing %s to %s\n", d.fileName, targetPath)
	return nil
}

func (d *GeminiDownloader) CleanFile(dir string) {
	if dir != "/" {
		os.RemoveAll(dir)
	}
	fmt.Printf("clean up file in %s\n", dir)
}
