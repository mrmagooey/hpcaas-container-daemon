package container

import "github.com/mitchellh/go-ps"
import "os/exec"
import "syscall"
import "github.com/mrmagooey/hpcaas-container-daemon/state"
import "os"
import "errors"
import "bytes"
import "time"

// creates a new exec cmd subprocess
// will return error if there is a problem creating the subprocess
// otherwise will spawn a goroutine that watches the subprocess
// check that we aren't already running
func ExecuteCode() error {
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
		envVars = append(envVars, key+"="+val)
	}
	cmd.Env = envVars
	// attach buffers to stderr stdout
	var out bytes.Buffer
	var err bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &err
	// start the code
	state.SetCodeStartedMethod(state.STARTED_BY_DAEMON)
	state.SetCodeState(state.CODE_RUNNING)
	if err := cmd.Start(); err != nil {
		state.SetCodeState(state.CODE_FAILED_TO_START)
		return errors.New("The code has failed to start")
	}
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

// The hpc code may start as a result of an ssh session starting it,
// rather than the daemon starting it (e.g. an MPI initiated start).
// This will watch the container process list for new processes
// and if one starts that matches the code at /hpcaas/code/<codename>
// this will start the watchCmd and update the daemons state to reflect this new process.
// This will be started in init() as a goroutine
func findProcess() {
	for {
		time.Sleep(1 * time.Second)
		if state.GetCodeState() == state.CODE_WAITING {
			// Get the process list
			procs, _ := ps.Processes()
			for _, psProc := range procs {
				if psProc.Executable() == state.GetCodeName() {
					pid := psProc.Pid()
					proc, _ := os.FindProcess(pid)
					state.SetCodeState(state.CODE_RUNNING)
					state.SetCodeStartedMethod(state.STARTED_EXTERNALLY)
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
			state.SetCodeState(state.CODE_STOPPED)
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

func init() {
	go findProcess()
}
