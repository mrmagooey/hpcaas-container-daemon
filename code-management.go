package main

import "os/exec"
import "syscall"
import "github.com/mrmagooey/hpcaas-container-daemon/state"
import "os"

// calls the hpc code and blocks until the HPC code finishes
// https://stackoverflow.com/questions/10385551/get-exit-code-go
// http://www.darrencoxall.com/golang/executing-commands-in-go/
func callHPCCode() {
	// get the name of the hpcaas code from the state
	codeName := state.GetCodeName()
	codePath := "/hpcaas/code/" + codeName
	if _, err := os.Stat(codePath); os.IsExist(err) {
		state.SetCodeState(state.CODE_MISSING)
		return
	}
	cmd := exec.Command(codePath)
	if err := cmd.Start(); err != nil {
		state.SetCodeState(state.CODE_ERROR)
		return
	}
	state.SetCodeState(state.CODE_RUNNING)
	// block on calling the code
	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// the code has died, there is a return code
			if _, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				state.SetCodeState(state.CODE_ERROR)
			}
		} else {
			// the code has died, but there is no return code
			state.SetCodeState(state.CODE_ERROR)
		}
	} else {
		state.SetCodeState(state.CODE_STOPPED)
	}
}
