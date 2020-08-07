package src

import (
	"fmt"
	"net"
	"path"

	"github.com/pkg/sftp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// HfSftp 结构
type HfSftp struct {
	sshClient *ssh.Client
	client    *sftp.Client
}

// NewHfSftp 创建sftp实例
func NewHfSftp(host string, port int, user string, pwd string) (*HfSftp, error) {
	sshConfig := &ssh.ClientConfig{
		User: user,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Auth: []ssh.AuthMethod{
			ssh.Password(pwd),
		},
		Timeout: 0,
	}

	//sshConfig.SetDefaults()
	var (
		e   HfSftp
		err error
	)
	if e.sshClient, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), sshConfig); err == nil {
		log.Info("Successfully connected to ssh server.")
		// open an SFTP session over an existing ssh connection.
		e.client, err = sftp.NewClient(e.sshClient)
	}
	return &e, err
}

// Close 关闭
func (e *HfSftp) Close() error {
	defer e.sshClient.Close()
	return e.client.Close()
}

// GetFileNames 取指定路径下的文件名
func (e *HfSftp) GetFileNames(remotePath string) ([]string, error) {
	var (
		files []string
		err   error
	)
	w := e.client.Walk(remotePath)
	for w.Step() {
		//w.SkipDir()
		if w.Err() != nil {
			continue
		}
		if w.Stat().IsDir() {
			continue
		}
		_, file := path.Split(w.Path())
		files = append(files[:], file)
	}
	return files, err
}

// GetFile 取远程文件
func (e *HfSftp) GetFile(fileName string) (*sftp.File, error) {
	return e.client.Open(fileName)
}
