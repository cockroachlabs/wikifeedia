package db

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

var haveCockroach bool

func init() {
	_, err := exec.LookPath("cockroach")
	haveCockroach = err == nil
}

func runCockroach() (_ *exec.Cmd, listenURL string, _ error) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, "", err
	}
	listenFileName := f.Name()
	f.Close()
	os.Remove(listenFileName)
	defer os.Remove(listenFileName)
	cmd := exec.Command("cockroach", "start", "--insecure",
		"--store=type=mem,size=1G",
		"--listen-addr=localhost:0",
		"--listening-url-file="+listenFileName)
	if err := cmd.Start(); err != nil {
		return nil, "", err
	}
	for {
		if _, err := os.Stat(listenFileName); os.IsNotExist(err) {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		break
	}
	contents, err := ioutil.ReadFile(listenFileName)
	if err != nil {
		return nil, "", err
	}
	return cmd, strings.TrimSpace(string(contents)), nil
}

func TestDB(t *testing.T) {
	if !haveCockroach {
		t.Skip("Don't have cockroach")
	}
	cockroach, pgUrl, err := runCockroach()
	if err != nil {
		t.Fatal(err)
	}
	defer cockroach.Process.Kill()
	_, err = New(pgUrl)
	if err != nil {
		t.Fatal(err)
	}
}
