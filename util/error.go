package util

import "errors"

var (
	UnexpectedNil = errors.New("unexpected nil value")

	NoSshClient   = errors.New("no ssh client")
	NoSftpSession = errors.New("no sftp session")

	UnknowSSHType = errors.New("unknow ssh type")
)
