package util

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func NewSSH_PW(user, password, host string, port int) (*ssh.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		err          error
	)
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))
	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	addr = fmt.Sprintf("%s:%d", host, port)
	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}
	return sshClient, err
}

func NewSSH_Key(user, privateKeyPath, host string, port int) (*ssh.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		err          error
	)

	privateKeyBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.PublicKeys(key))

	clientConfig = &ssh.ClientConfig{
		User:            user,
		Auth:            auth,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr = fmt.Sprintf("%s:%d", host, port)
	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	return sshClient, nil
}

func NewSftpClient(sshClient *ssh.Client) (*sftp.Client, error) {
	var (
		sftpClient *sftp.Client
		err        error
	)
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		return nil, err
	}
	return sftpClient, nil
}

// TODO(Benevor):timeout interval
func UploadFile(ip string, localFilePath string, remoteDir string, sftpClient *sftp.Client) error {
	fmt.Printf("start uploading %s to %s:%s \n", localFilePath, ip, remoteDir)
	if sftpClient == nil {
		return NoSftpSession
	}
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		fmt.Printf("open %s failed! %v\n", localFilePath, err)
		return err
	}
	defer srcFile.Close()
	// create target file on remote host
	var remoteFileName = path.Base(localFilePath)
	dstFile, err := sftpClient.Create(path.Join(remoteDir, remoteFileName))
	if err != nil {
		fmt.Printf("create %s:%s failed! %v\n", ip, path.Join(remoteDir, remoteFileName), err)
		return err
	}
	defer dstFile.Close()
	if _, err = dstFile.ReadFrom(srcFile); err != nil {
		fmt.Printf("%s:%s read from %s failed! %v\n", ip, path.Join(remoteDir, remoteFileName), localFilePath, err)
		return err
	}
	fmt.Printf("finish uploading %s to %s:%s \n", localFilePath, ip, remoteDir)
	return nil
}

func NewSshSession(sshClient *ssh.Client) (*ssh.Session, error) {
	var (
		sshSession *ssh.Session
		err        error
	)
	if sshSession, err = sshClient.NewSession(); err != nil {
		return nil, err
	}
	return sshSession, nil
}
