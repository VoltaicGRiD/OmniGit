package main

import (
	"os/exec"
	"fmt"
	"io"
)

func RunExecCommand(command string, arguments []string) ([]byte, []byte) {
	cmd := exec.Command(command, arguments...)
	stderr, err := cmd.StderrPipe()
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		WriteErr(err.Error())
	}

	// TODO Naming is hard
	var errout []byte
	var outout []byte

	if err := cmd.Start(); err != nil {
		errout, _ = io.ReadAll(stderr)
		WriteErr(fmt.Sprintf("%s", errout))
		WriteErr(err.Error())
	}

	outout, _ = io.ReadAll(stdout)

	return outout, errout
}
