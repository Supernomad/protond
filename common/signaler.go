// Copyright (c) 2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/protond/blob/master/LICENSE

package common

import (
	"errors"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

// Signaler struct used to manage os and user signals to the protond process.
type Signaler struct {
	log    *Logger
	config *Config

	fds     []int
	env     map[string]string
	signals chan os.Signal
}

func (sig *Signaler) fork(exec bool, files []uintptr) (int, error) {
	if exec {
		return syscall.ForkExec(os.Args[0], os.Args, &syscall.ProcAttr{Env: os.Environ(), Files: files})
	}
	return -1, nil
}

func (sig *Signaler) reload(exec bool) error {
	sig.log.Info.Println("[MAIN]", "Received reload signal from user. Reloading process...")

	files := make([]uintptr, 3+len(sig.fds))
	files[0] = os.Stdin.Fd()
	files[1] = os.Stdout.Fd()
	files[2] = os.Stderr.Fd()

	for i := 0; i < len(sig.fds); i++ {
		files[3+i] = uintptr(sig.fds[i])
	}

	for k, v := range sig.env {
		os.Setenv(k, v)
	}

	pid, err := sig.fork(exec, files)
	if err != nil {
		return errors.New("error execing new instance of protond during reload: " + err.Error())
	}

	err = ioutil.WriteFile(sig.config.PidFile, []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		return errors.New("error the new pid for the new instance of protond during reload: " + err.Error())
	}
	return nil
}

func (sig *Signaler) terminate(exec bool) error {
	sig.log.Info.Println("[MAIN]", "Received termination signal from user. Terminating process.")
	return nil
}

// Wait for a configured os or user signal to be passed to the protond process.
func (sig *Signaler) Wait(exec bool) error {
	s := <-sig.signals
	switch s {
	case syscall.SIGHUP:
		return sig.reload(exec)
	case syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT:
		return sig.terminate(exec)
	default:
		return errors.New("build error received undefined signal")
	}
}

// NewSignaler generates a new Signaler object, which will watch for new os and user signals passed to the protond process.
func NewSignaler(log *Logger, config *Config, fds []int, env map[string]string) *Signaler {
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	return &Signaler{
		log:     log,
		config:  config,
		fds:     fds,
		env:     env,
		signals: signals,
	}
}
