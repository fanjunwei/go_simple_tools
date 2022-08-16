package main

import (
	"os"
	"strings"
	"syscall"
)
import (
	"os/exec"
)

func main() {
	strArg := strings.Join(os.Args[1:], " ")
	cmd := exec.Command("/bin/bash", "-c", strArg)
	//cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setctty: true, Setpgid: true, Credential: &syscall.Credential{Uid: 0, Gid: 0}}
	cmd.Run()
}
