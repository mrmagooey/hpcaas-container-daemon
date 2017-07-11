package container

import "os/exec"
import "syscall"
import "github.com/mrmagooey/hpcaas-container-daemon/state"
import "os"
import "errors"

// the process command struct
var cmd = &exec.Cmd{}

// calls the hpc code and blocks until the HPC code finishes
// https://stackoverflow.com/questions/10385551/get-exit-code-go
// http://www.darrencoxall.com/golang/executing-commands-in-go/
func ExecuteCode() error {
	// check that we aren't already running
	if state.GetCodeState() != state.CODE_WAITING {
		return errors.New("Code cannot be started")
	}
	// get the name of the hpcaas code from the state
	codeName := state.GetCodeName()
	codePath := "/hpcaas/code/" + codeName
	if _, err := os.Stat(codePath); os.IsExist(err) {
		state.SetCodeState(state.CODE_MISSING)
		return errors.New("Code executable is missing")
	}
	cmd = exec.Command(codePath)
	// start the code
	if err := cmd.Start(); err != nil {
		state.SetCodeState(state.CODE_ERROR)
		return errors.New("The code has failed to start")
	}
	state.SetCodeState(state.CODE_RUNNING)
	// start the watcher
	go watchCmd()
	return nil
}

func watchCmd() {
	// block on calling the code
	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// the code has died, there is a return code
			if _, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				// if we killed the code don't change the state to error
				if state.GetCodeState() != state.CODE_KILLED {
					state.SetCodeState(state.CODE_ERROR)
				}
			}
		} else {
			// the code has died, but there is no return code
			if state.GetCodeState() != state.CODE_KILLED {
				state.SetCodeState(state.CODE_ERROR)
			}
		}
	} else {
		// the code has finished with a return code of 0
		state.SetCodeState(state.CODE_STOPPED)
	}
}

func KillCode() error {
	// set the state to killed, needs to be done before we actually try to kill it
	state.SetCodeState(state.CODE_KILLED)
	if err := cmd.Process.Kill(); err != nil {
		// if we failed to kill the process, reset the state
		state.SetCodeState(state.CODE_RUNNING)
		return errors.New("Failed to kill process")
	}
	return nil
}
