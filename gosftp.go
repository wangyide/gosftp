package gosftp

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// readPrivateKey readPrivateKey
func readPrivateKey(path string) (ssh.Signer, error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		log.Print(err)
		return nil, err
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return ssh.ParsePrivateKey(b)
}

// ReciveFile file download
func ReciveFile(rmthost, port, srcPath, dstPath, filename, remoteUser, keyPath string) (int64, error) {
	var (
		copied int64
		err    error
	)
	key, err := readPrivateKey(keyPath)
	if err != nil {
		log.Print(err)
		return copied, err
	}
	// Define the Client Config
	config := &ssh.ClientConfig{
		User: remoteUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", rmthost, port), config)
	if err != nil {
		log.Print(err)
		return copied, err
	}
	// open an SFTP session over an existing ssh connection.
	sftp, err := sftp.NewClient(client)
	if err != nil {
		log.Print(err)
		return copied, err
	}
	defer sftp.Close()

	// Open the source file
	srcFile, err := sftp.Open(srcPath)
	if err != nil {
		log.Print(err)
		return copied, err
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(sftp.Join(dstPath, filename))
	if err != nil {
		log.Print(err)
		return copied, err
	}
	defer dstFile.Close()

	// Copy the file
	copied, err = srcFile.WriteTo(dstFile)

	return copied, err
}

// ReciveFolder folder download
func ReciveFolder(rmthost, port, srcPath, dstPath, remoteUser, keyPath string) (int, error) {
	var (
		copied int
		err    error
	)
	key, err := readPrivateKey(keyPath)
	if err != nil {
		log.Print(err)
		return copied, err
	}
	// Define the Client Config
	config := &ssh.ClientConfig{
		User: remoteUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", rmthost, port), config)
	if err != nil {
		log.Print(err)
		return copied, err
	}
	// open an SFTP session over an existing ssh connection.
	sftp, err := sftp.NewClient(client)
	if err != nil {
		log.Print(err)
		return copied, err
	}
	defer sftp.Close()

	// Open the source file
	srcFile, err := sftp.Open(srcPath)
	if err != nil {
		log.Print(err)
		return copied, err
	}
	defer srcFile.Close()

	files, err := sftp.ReadDir(srcPath)
	if err != nil {
		log.Print(err)
		return copied, err
	}
	os.Mkdir(dstPath, os.ModePerm)
	dirinfo, err := os.Stat(dstPath)
	if err != nil || !dirinfo.IsDir() {
		log.Print(err)
		return copied, err
	}

	for _, file := range files {
		// Open the source file
		srcfilepath := sftp.Join(srcPath, file.Name())
		srcFile, err := sftp.Open(srcfilepath)
		if err != nil {
			log.Printf("Open remote file %v error:%v", srcfilepath, err)
			return copied, err
		}
		defer srcFile.Close()

		// Create the destination file
		dstfilepath := sftp.Join(dstPath, file.Name())
		dstFile, err := os.Create(dstfilepath)
		if err != nil {
			log.Printf("Create local file %v error:%v", dstfilepath, err)
			return copied, err
		}
		os.Chmod(dstfilepath, file.Mode())
		defer dstFile.Close()

		// Copy the file
		_, err = srcFile.WriteTo(dstFile)
		if err != nil {
			log.Print(err)
			return copied, err
		}
		copied++
	}

	return copied, err
}
