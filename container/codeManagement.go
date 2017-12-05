package container

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/mitchellh/go-ps"
	"github.com/mrmagooey/hpcaas-common"
	"github.com/mrmagooey/hpcaas-container-daemon/state"
)

// ExecuteCode creates a new exec cmd subprocess
// will return error if there is a problem creating the subprocess
// otherwise will spawn a goroutine that watches the subprocess
// check that we aren't already running
func ExecuteCode() error {
	if state.GetCodeStatus() != common.CodeWaitingStatus {
		return errors.New("Code already started")
	}
	// get hpcaas code info from state
	codeName := state.GetCodeName()
	codeArgs := state.GetCodeArguments()
	codePath := "/hpcaas/code/" + codeName
	if _, err := os.Stat(codePath); err != nil {
		state.SetCodeStatus(common.CodeMissingStatus)
		return errors.New("Code executable is missing")
	}
	cmd := exec.Command(codePath, codeArgs...)
	// get the environment variables
	codeParams := state.GetCodeParams()
	var envVars []string
	for key, val := range codeParams {
		envVars = append(envVars, key+"="+val)
	}
	cmd.Env = envVars
	// attach buffers to stderr stdout
	var out bytes.Buffer
	var err bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &err
	// start the code
	state.SetCodeStartedMethod(common.StartedByDaemonStatus)
	state.SetCodeStatus(common.CodeRunningStatus)
	if err := cmd.Start(); err != nil {
		state.SetCodeStatus(common.CodeFailedToStartStatus)
		return errors.New("The code has failed to start")
	}
	state.SetCodePID(cmd.Process.Pid)
	// start two goroutines, one to watch the running code
	// the other to listen for a kill signal
	go watchCmd(cmd, &out, &err)
	return nil
}

// KillCode send the kill signal
func KillCode() error {
	if s := state.GetCodeStatus(); s != common.CodeRunningStatus {
		return errors.New("No process currently running")
	}
	state.SetCodeStatus(common.CodeKilledStatus)
	proc, err := os.FindProcess(state.GetCodePID())
	if err != nil {
		log.Println(err.Error())
		state.SetCodeStatus(common.CodeFailedToKillStatus)
	}
	// extra check that the process is running
	err = proc.Signal(syscall.Signal(0))
	if err != nil {
		// TODO
		// process has died
		// watchcmd should have updated everything
		return nil
	}
	// tell the process to terminate
	err = proc.Signal(syscall.SIGTERM)
	if err != nil {
		log.Println(err.Error())
		state.SetCodeStatus(common.CodeFailedToKillStatus)
	}
	return nil
}

// The hpc code may start as a result of an ssh session starting it,
// rather than the daemon starting it (e.g. an MPI initiated start).
// This will watch the container process list for new processes
// and if one starts that matches the code at /hpcaas/code/<codename>
// this will start the watchCmd and update the daemons state to reflect this new process.
// This will be started in init() as a goroutine
func findProcess() {
	for {
		time.Sleep(1 * time.Second)
		if state.GetCodeStatus() == common.CodeWaitingStatus {
			// Get the process list
			procs, _ := ps.Processes()
			for _, psProc := range procs {
				if psProc.Executable() == state.GetCodeName() {
					pid := psProc.Pid()
					proc, _ := os.FindProcess(pid)
					state.SetCodeStatus(common.CodeRunningStatus)
					state.SetCodeStartedMethod(common.StartedExternallyStatus)
					state.SetCodePID(pid)
					go watchProc(proc)
				}
			}
		}
	}
}

func watchProc(proc *os.Process) {
	for {
		time.Sleep(1 * time.Second)
		err := proc.Signal(syscall.Signal(0))
		if err != nil {
			// process has died, need to update things
			state.SetCodeStatus(common.CodeStoppedStatus)
		}
	}
}

// watches the cmd and takes action upon its death
// blocks until the HPC code finishes
// https://stackoverflow.com/questions/10385551/get-exit-code-go
// http://www.darrencoxall.com/golang/executing-commands-in-go/
func watchCmd(cmd *exec.Cmd, out *bytes.Buffer, err *bytes.Buffer) {
	// block on calling the code
	if err := cmd.Wait(); err != nil {
		// the code has died
		if exiterr, ok := err.(*exec.ExitError); ok {
			// there is a return code
			if _, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				// if we killed the code don't change the state to error
				if state.GetCodeStatus() != common.CodeKilledStatus {
					state.SetCodeStatus(common.CodeErrorStatus)
					log.Println(err.Error())
				}
			}
		} else {
			// the code has died, but there is no return code (?)
			if state.GetCodeStatus() != common.CodeKilledStatus {
				state.SetCodeStatus(common.CodeErrorStatus)
			}
		}
	} else {
		// the code has finished with a return code of 0
		state.SetCodeStatus(common.CodeStoppedStatus)
	}
	state.SetCodeStdout(out.String())
	state.SetCodeStderr(err.String())
}

func init() {
	go findProcess()
}
