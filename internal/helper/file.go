package helper

import (
	"bytes"
	"errors"
	"os/exec"
)

func Exec(name string, arg ...string) (output string, err error) {
	cmd := exec.Command(name, arg...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	er := cmd.Run()
	stdOutput, errOutput := string(stdout.Bytes()), string(stderr.Bytes())
	if er != nil {
		err = errors.New(errOutput)
		return
	}

	output = stdOutput
	return
}
