package tkInfra

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"os/user"
	"slices"
	"strconv"
	"strings"
	"syscall"
)

type Shell struct {
	runtimeSettings ShellSettings
}

type ShellSettings struct {
	Command                         string
	Args                            []string
	ShouldUseSubShell               bool
	ShouldDisableTimeoutHardLimit   bool
	ShouldIgnoreUsernameLookupError bool
	Username                        string
	WorkingDirectory                string
	ExecutionTimeoutSecs            uint64
	Envs                            []string
	StdoutFilePath                  string
	StderrFilePath                  string
}

func NewShell(settings ShellSettings) Shell {
	return Shell{runtimeSettings: settings}
}

type ShellError struct {
	StdErr   string `json:"stdErr"`
	ExitCode int    `json:"exitCode"`
}

func (e *ShellError) Error() string {
	jsonError, _ := json.Marshal(e)
	return string(jsonError)
}

func (shell Shell) sysCallCredentialsFactory() (*syscall.Credential, error) {
	userStruct, err := user.Lookup(shell.runtimeSettings.Username)
	if err != nil {
		return nil, err
	}
	userId, err := strconv.Atoi(userStruct.Uid)
	if err != nil {
		return nil, err
	}
	groupId, err := strconv.Atoi(userStruct.Gid)
	if err != nil {
		return nil, err
	}

	return &syscall.Credential{
		Uid: uint32(userId),
		Gid: uint32(groupId),
	}, nil
}

type preparedExec struct {
	ExecCmd           *exec.Cmd
	StdoutBytesBuffer *bytes.Buffer
	StdoutFileHandler *os.File
	StderrBytesBuffer *bytes.Buffer
	StderrFileHandler *os.File
	Err               error
}

func (shell Shell) prepareExec() preparedExec {
	if shell.runtimeSettings.ShouldUseSubShell {
		subShellCmd := shell.runtimeSettings.Command + " " +
			strings.Join(shell.runtimeSettings.Args, " ")
		subShellArgs := []string{"-c", "source /etc/profile; " + subShellCmd}
		shell.runtimeSettings.Command = "bash"
		shell.runtimeSettings.Args = subShellArgs
	}

	timeoutSecsDefault := uint64(1800)
	if shell.runtimeSettings.ExecutionTimeoutSecs == 0 {
		shell.runtimeSettings.ExecutionTimeoutSecs = timeoutSecsDefault
	}

	timeoutSecsHardLimit := uint64(3600)
	if shell.runtimeSettings.ExecutionTimeoutSecs > timeoutSecsHardLimit &&
		!shell.runtimeSettings.ShouldDisableTimeoutHardLimit {
		shell.runtimeSettings.ExecutionTimeoutSecs = timeoutSecsHardLimit
	}

	timeoutSecsStr := strconv.FormatUint(shell.runtimeSettings.ExecutionTimeoutSecs, 10)

	timeoutArgs := []string{timeoutSecsStr, shell.runtimeSettings.Command}
	timeoutArgs = slices.Concat(timeoutArgs, shell.runtimeSettings.Args)
	shell.runtimeSettings.Command = "timeout"
	shell.runtimeSettings.Args = timeoutArgs

	execCmd := exec.Command(
		shell.runtimeSettings.Command, shell.runtimeSettings.Args...,
	)
	if shell.runtimeSettings.Username != "" {
		sysCallCredentials, err := shell.sysCallCredentialsFactory()
		if err != nil && !shell.runtimeSettings.ShouldIgnoreUsernameLookupError {
			return preparedExec{Err: err}
		}
		if err == nil {
			execCmd.SysProcAttr = &syscall.SysProcAttr{Credential: sysCallCredentials}
		}
	}

	if shell.runtimeSettings.WorkingDirectory != "" {
		execCmd.Dir = shell.runtimeSettings.WorkingDirectory
	}

	var stdoutBytesBuffer bytes.Buffer
	execCmd.Stdout = &stdoutBytesBuffer
	if shell.runtimeSettings.StdoutFilePath != "" {
		stdoutFileHandler, err := os.Create(shell.runtimeSettings.StdoutFilePath)
		if err != nil {
			return preparedExec{Err: err}
		}
		execCmd.Stdout = stdoutFileHandler
	}

	var stderrBytesBuffer bytes.Buffer
	execCmd.Stderr = &stderrBytesBuffer
	if shell.runtimeSettings.StderrFilePath != "" {
		stderrFileHandler, err := os.Create(shell.runtimeSettings.StderrFilePath)
		if err != nil {
			return preparedExec{Err: err}
		}
		execCmd.Stderr = stderrFileHandler
	}

	execCmd.Env = append(execCmd.Environ(), "DEBIAN_FRONTEND=noninteractive")
	execCmd.Env = slices.Concat(execCmd.Env, shell.runtimeSettings.Envs)

	return preparedExec{
		ExecCmd:           execCmd,
		StdoutBytesBuffer: &stdoutBytesBuffer,
		StderrBytesBuffer: &stderrBytesBuffer,
	}
}

func (shell Shell) Run() (stdoutStr string, err error) {
	preparedExec := shell.prepareExec()
	if preparedExec.Err != nil {
		return stdoutStr, preparedExec.Err
	}

	preparedExec.Err = preparedExec.ExecCmd.Run()
	if preparedExec.StdoutFileHandler != nil {
		preparedExec.StdoutFileHandler.Close()
	}
	if preparedExec.StderrFileHandler != nil {
		preparedExec.StderrFileHandler.Close()
	}

	if preparedExec.StdoutBytesBuffer != nil {
		stdoutStr = strings.TrimSpace(preparedExec.StdoutBytesBuffer.String())
	}
	if preparedExec.Err == nil {
		return stdoutStr, nil
	}

	if exitErr, assertOk := preparedExec.Err.(*exec.ExitError); assertOk {
		stdErrStr := preparedExec.StderrBytesBuffer.String()
		if exitErr.ExitCode() == 124 {
			stdErrStr = "CommandDeadlineExceeded"
		}

		return stdoutStr, &ShellError{
			StdErr:   stdErrStr,
			ExitCode: exitErr.ExitCode(),
		}
	}

	return stdoutStr, preparedExec.Err
}
