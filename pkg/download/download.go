package download

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"openGemini-UP/util"
	"os"
	"path/filepath"
	"strings"
	"time"
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
		website:     util.Download_web,
		version:     ops.Version,
		typ:         "-" + ops.Os + "-" + ops.Arch + util.Download_pkg_suffix,
		destination: util.Download_dst,
		timeout:     util.Download_timeout,
	}
}

func (d *GeminiDownloader) setVersion(v string) {
	d.version = v
}

func (d *GeminiDownloader) setType(t string) {
	d.typ = t
}

func (d *GeminiDownloader) setDestination(dst string) {
	d.destination = dst
}

func (d *GeminiDownloader) setTimeout(t time.Duration) {
	d.timeout = t
}

func (d *GeminiDownloader) spliceUrl() error {
	if d.website == "" {
		d.website = util.Download_web
	}

	if d.version == "" {
		d.version = util.Download_default_version
	}

	if err := d.checkVersion(); err != nil {
		return err
	}

	d.Url = d.website + "/" + d.version + "/" + util.Download_fill_char + d.version[1:] + d.typ
	return nil
}

func (d *GeminiDownloader) checkVersion() error {
	return nil
}

func (d *GeminiDownloader) Run() error {
	if _, err := os.Stat(util.Download_dst); os.IsNotExist(err) {
		errDir := os.MkdirAll(util.Download_dst, 0755)
		if errDir != nil {
			return errDir
		}
	}

	if d.isMissing() { // check whether need to download the files
		dir := filepath.Join(d.destination, d.version)
		if err := d.spliceUrl(); err != nil {
			d.CleanFile(dir)
			return err
		}

		if err := d.downloadFile(); err != nil {
			d.CleanFile(dir)
			return err
		}

		if err := d.decompressFile(); err != nil {
			d.CleanFile(dir)
			return err
		}
	}

	return nil
}

func (d *GeminiDownloader) isMissing() bool {
	dir := filepath.Join(d.destination, d.version)
	_, err := os.Stat(dir)
	return os.IsNotExist(err)
}

func (d *GeminiDownloader) downloadFile() error {
	dir := filepath.Join(d.destination, d.version)
	fmt.Printf("start downloading file from %s to %s\n", d.Url, dir)

	var client *http.Client
	// get HTTP_PROXY and parse
	httpProxy := os.Getenv("HTTP_PROXY")
	if httpProxy != "" {
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

	// create local file
	d.CleanFile(dir)
	if err = os.Mkdir(dir, 0755); err != nil {
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

	// open .tar.gz file
	file, err := os.Open(d.fileName)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file.Close()

	// new gzip.Reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer gzReader.Close()

	// new tar.Reader
	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			return err
		}

		targetFilePath := filepath.Join(targetPath, header.Name)
		// If it is a directory, create the corresponding directory
		if header.Typeflag == tar.TypeDir {
			err = os.MkdirAll(targetFilePath, header.FileInfo().Mode())
			if err != nil {
				fmt.Println(err)
				return err
			}
			continue
		}
		// If it is a file, create and copy the contents of the file
		if header.Typeflag == tar.TypeReg {
			targetFile, err := os.OpenFile(targetFilePath, os.O_CREATE|os.O_RDWR, header.FileInfo().Mode())
			if err != nil {
				fmt.Println(err)
				return err
			}
			defer targetFile.Close()

			_, err = io.Copy(targetFile, tarReader)
			if err != nil {
				fmt.Println(err)
				return err
			}
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
