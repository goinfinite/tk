package tkInfra

import (
	"bytes"
	"encoding/json"
	"errors"
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

func (shell Shell) prepareCmdExecutor() (*exec.Cmd, *bytes.Buffer, *bytes.Buffer) {
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

	cmdExecutor := exec.Command(
		shell.runtimeSettings.Command, shell.runtimeSettings.Args...,
	)
	if shell.runtimeSettings.Username != "" {
		sysCallCredentials, err := shell.sysCallCredentialsFactory()
		if err == nil {
			cmdExecutor.SysProcAttr = &syscall.SysProcAttr{Credential: sysCallCredentials}
		}
		if err != nil && !shell.runtimeSettings.ShouldIgnoreUsernameLookupError {
			return nil, nil, nil
		}
	}

	if shell.runtimeSettings.WorkingDirectory != "" {
		cmdExecutor.Dir = shell.runtimeSettings.WorkingDirectory
	}

	var stdoutBytesBuffer, stderrBytesBuffer bytes.Buffer
	cmdExecutor.Stdout = &stdoutBytesBuffer
	cmdExecutor.Stderr = &stderrBytesBuffer

	cmdExecutor.Env = append(cmdExecutor.Environ(), "DEBIAN_FRONTEND=noninteractive")
	cmdExecutor.Env = slices.Concat(cmdExecutor.Env, shell.runtimeSettings.Envs)

	return cmdExecutor, &stdoutBytesBuffer, &stderrBytesBuffer
}

func (shell Shell) Run() (string, error) {
	cmdExecutor, stdoutBytesBuffer, stderrBytesBuffer := shell.prepareCmdExecutor()
	if cmdExecutor == nil {
		return "", errors.New("UsernameLookupError")
	}

	err := cmdExecutor.Run()
	stdoutStr := strings.TrimSpace(stdoutBytesBuffer.String())
	if err == nil {
		return stdoutStr, nil
	}

	if exitErr, assertOk := err.(*exec.ExitError); assertOk {
		stdErrStr := stderrBytesBuffer.String()
		if exitErr.ExitCode() == 124 {
			stdErrStr = "CommandDeadlineExceeded"
		}

		return stdoutStr, &ShellError{
			StdErr:   stdErrStr,
			ExitCode: exitErr.ExitCode(),
		}
	}

	return stdoutStr, err
}
