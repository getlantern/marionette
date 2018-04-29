// +build python

package fte_test

import (
	"io/ioutil"
	"os"
	"os/exec"
)

// RunPython runs a Python script with the given arguments.
func RunPython(program string, args ...string) (stdout []byte, err error) {
	filename, err := WriteTempFile([]byte(program), 0700)
	if err != nil {
		return nil, err
	}
	defer os.Remove(filename)

	cmd := exec.Command("python2", append([]string{filename}, args...)...)
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

// WriteTempFile writes data to a temporary path. Returns the filename.
func WriteTempFile(data []byte, mode os.FileMode) (filename string, err error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	} else if err := f.Close(); err != nil {
		return "", err
	} else if err := os.Remove(f.Name()); err != nil {
		return "", err
	}

	filename = f.Name() + ".py"
	if err := ioutil.WriteFile(filename, data, mode); err != nil {
		return "", err
	}
	return filename, nil
}
