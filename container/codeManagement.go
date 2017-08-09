package container

import "os/exec"
import "syscall"
import "github.com/mrmagooey/hpcaas-container-daemon/state"
import "os"
import "errors"
import "bytes"

func ExecuteCode() error {
	// creates a new exec cmd subprocess
	// will return error if there is a problem creating the subprocess
	// otherwise will spawn a goroutine that watches the subprocess
	// check that we aren't already running
	if state.GetCodeState() != state.CODE_WAITING {
		return errors.New("Code already started")
	}
	// get hpcaas code info from state
	codeName := state.GetCodeName()
	codeArgs := state.GetCodeArguments()
	codePath := "/hpcaas/code/" + codeName
	if _, err := os.Stat(codePath); err != nil {
		state.SetCodeState(state.CODE_MISSING)
		return errors.New("Code executable is missing")
	}
	cmd := exec.Command(codePath, codeArgs...)
	// get the environment variables
	codeParams := state.GetCodeParams()
	var envVars []string
	for key, val := range codeParams {
		envVars = append(envVars, key+"="+val.(string))
	}
	cmd.Env = envVars
	// attach buffers to stderr stdout
	var out bytes.Buffer
	var err bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &err
	// start the code
	if err := cmd.Start(); err != nil {
		state.SetCodeState(state.CODE_FAILED_TO_START)
		return errors.New("The code has failed to start")
	}
	state.SetCodeState(state.CODE_RUNNING)
	state.SetCodePID(cmd.Process.Pid)
	// start two goroutines, one to watch the running code
	// the other to listen for a kill signal
	go watchCmd(cmd, &out, &err)
	return nil
}

// send the kill signal
func KillCode() error {
	if s := state.GetCodeState(); s != state.CODE_RUNNING {
		return errors.New("No process currently running")
	}
	state.SetCodeState(state.CODE_KILLED)
	proc, err := os.FindProcess(state.GetCodePID())
	// extra check that the process is running
	err = proc.Signal(syscall.Signal(0))
	if err != nil {
		// process has died
		// watchcmd should have updated everything
		return nil
	} else {
		// tell the process to terminate
		err := proc.Signal(syscall.SIGTERM)
		if err != nil {
			state.AddErrorMessage(err.Error())
			state.SetCodeState(state.CODE_FAILED_TO_KILL)
		}
	}
	return nil
}

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
				if state.GetCodeState() != state.CODE_KILLED {
					state.SetCodeState(state.CODE_ERROR)
					state.AddErrorMessage(err.Error())
				}
			}
		} else {
			// the code has died, but there is no return code (?)
			if state.GetCodeState() != state.CODE_KILLED {
				state.SetCodeState(state.CODE_ERROR)
			}
		}
	} else {
		// the code has finished with a return code of 0
		state.SetCodeState(state.CODE_STOPPED)
	}
	state.SetCodeStdout(out.String())
	state.SetCodeStderr(err.String())
}
