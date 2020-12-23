package main

// main.go

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)
func sshconnect(user, password, host string, port int) (*ssh.Session, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		session      *ssh.Session
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create session
	if session, err = client.NewSession(); err != nil {
		return nil, err
	}

	return session, nil
}

func sftpconnect(user, password, host string, port int) (*sftp.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		sftpClient   *sftp.Client
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
		//这个是问你要不要验证远程主机，以保证安全性。这里不验证
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create sftp client
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		return nil, err
	}

	return sftpClient, nil
}


//单个copy
func scpCopy(localFilePath string, remoteDir string, user string,userPwd string, ip string) error {
	fmt.Println("正在上传.....：",ip)
	var (
		sftpClient *sftp.Client
		err        error
	)
	// 这里换成实际的 SSH 连接的 用户名，密码，主机名或IP，SSH端口
	sftpClient, err = sftpconnect(user, userPwd, ip, 22)
	if err != nil {
		fmt.Println("传输错误:", err)
		return err
	}
	defer sftpClient.Close()
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		fmt.Println("传输错误:", err)
		return err
	}
	defer srcFile.Close()

	var remoteFileName = path.Base(localFilePath)
	dstFile, err := sftpClient.Create(path.Join(remoteDir, remoteFileName))
	if err != nil {
		fmt.Println("传输错误:", err)
		return err
	}
	defer dstFile.Close()

	buf := make([]byte, 1024)
	for {
		n, _ := srcFile.Read(buf)
		if n == 0 {
			break
		}
		dstFile.Write(buf[0:n])
	}
	fmt.Println("上传完成：",ip)
	return nil
}


//读取key=value类型的配置文件
func initConfig(path string) map[string]string {
	config := make(map[string]string)

	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		panic(err)
	}

	r := bufio.NewReader(f)
	for {
		b, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		s := strings.TrimSpace(string(b))
		index := strings.Index(s, "=")
		if index < 0 {
			continue
		}
		key := strings.TrimSpace(s[:index])
		if len(key) == 0 {
			continue
		}
		value := strings.TrimSpace(s[index+1:])
		if len(value) == 0 {
			continue
		}
		config[key] = value
	}
	return config
}

func getInput() string {
	//使用os.Stdin开启输入流
	//函数原型 func NewReader(rd io.Reader) *Reader
	//NewReader创建一个具有默认大小缓冲、从r读取的*Reader 结构见官方文档
	in := bufio.NewReader(os.Stdin)
	//in.ReadLine函数具有三个返回值 []byte bool error
	//分别为读取到的信息 是否数据太长导致缓冲区溢出 是否读取失败
	str, _, err := in.ReadLine()
	if err != nil {
		return err.Error()
	}
	return string(str)
}

func main() {

	fmt.Println("------------------Copyright ©2021 www.wgstart.com -------------------------------------")
	fmt.Print("需传输的本地文件完整路径:")
	localFilePath := getInput()
	fmt.Println(localFilePath)

	fmt.Print("传输到目标主机的远程存贮路径:")
	remoteDir := getInput()
	fmt.Println(remoteDir)



	hostMaps := initConfig("./host.properties")

	for k, v := range hostMaps {
		if strings.Contains(k,"#"){
			continue
		}
		if v == "" {
			fmt.Println("value is null：", k)
			continue
		}
		user := strings.Split(v,"//")[0]
		userpwd := strings.Split(v,"//")[1]
		remoteDir =  strings.ReplaceAll(remoteDir," ","")
		localFilePath =  strings.ReplaceAll(localFilePath," ","")
		scpCopy(localFilePath, remoteDir,user, userpwd,k)
	}

	fmt.Println("上传结束")

	time.Sleep(100 * time.Hour)

}