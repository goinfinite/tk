package tkInfra

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

const (
	ShellExecutionTimeoutDefaultSecs   uint64 = 1800
	ShellExecutionTimeoutHardLimitSecs uint64 = 3600
	ShellExecutionTimeoutGraceSecs     uint64 = 10
	ShellCommandTimeoutExitCode        int    = 124
)

// IsStdoutTerminal is the single interactivity check shared by the CLI logger
// and response renderer. Both honor one contract: logs go to stderr, stdout
// carries the JSON response only, and a human at a terminal gets richer
// formatting. Both sides must agree or the response channel corrupts.
func IsStdoutTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

type Shell struct {
	runtimeSettings ShellSettings
}

type ShellSettings struct {
	Command                         string
	Args                            []string
	ShouldUseSubShell               bool
	ShouldUseCleanEnv               bool
	ShouldDisableTimeoutHardLimit   bool
	ShouldIgnoreUsernameLookupError bool
	Username                        string
	WorkingDirectory                string
	ExecutionTimeoutSecs            uint64
	Envs                            []string
	StdoutFilePath                  string
	StderrFilePath                  string
}

// NewShell returns a value, not a pointer. Run copies the settings per call,
// so the sub-shell rewrite during preparation never compounds. A pointer
// receiver was rejected: it would leak that rewrite across calls and
// double-wrap the command on the second Run.
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

func (shell Shell) sysCallCredentialsFactory() (
	*syscall.Credential, *user.User, error,
) {
	userStruct, err := user.Lookup(shell.runtimeSettings.Username)
	if err != nil {
		return nil, nil, err
	}
	userId, err := strconv.Atoi(userStruct.Uid)
	if err != nil {
		return nil, nil, err
	}
	groupId, err := strconv.Atoi(userStruct.Gid)
	if err != nil {
		return nil, nil, err
	}

	return &syscall.Credential{
		Uid: uint32(userId),
		Gid: uint32(groupId),
	}, userStruct, nil
}

func (shell Shell) childEnvironmentBuilder(
	execCmd *exec.Cmd,
	targetUserPtr *user.User,
) []string {
	const debianFrontendEnv = "DEBIAN_FRONTEND=noninteractive"
	const standardSystemPath = "/usr/local/sbin:/usr/local/bin:" +
		"/usr/sbin:/usr/bin:/sbin:/bin"

	if !shell.runtimeSettings.ShouldUseCleanEnv {
		inheritedEnv := append(execCmd.Environ(), debianFrontendEnv)
		return slices.Concat(inheritedEnv, shell.runtimeSettings.Envs)
	}

	userHome := os.Getenv("HOME")
	if targetUserPtr != nil {
		userHome = targetUserPtr.HomeDir
	}

	childPath := standardSystemPath
	if userHome != "" {
		childPath = userHome + "/.local/bin:" + standardSystemPath
	}

	cleanEnv := []string{"PATH=" + childPath, debianFrontendEnv}
	if userHome != "" {
		cleanEnv = append(cleanEnv, "HOME="+userHome)
	}
	if execCmd.Dir != "" {
		cleanEnv = append(cleanEnv, "PWD="+execCmd.Dir)
	}

	return slices.Concat(cleanEnv, shell.runtimeSettings.Envs)
}

func (shell Shell) closeFileHandler(fileHandlerPtr *os.File) error {
	if fileHandlerPtr == nil {
		return nil
	}

	return fileHandlerPtr.Close()
}

type executionPlan struct {
	ExecCmd           *exec.Cmd
	StdoutBytesBuffer *bytes.Buffer
	StdoutFileHandler *os.File
	StderrBytesBuffer *bytes.Buffer
	StderrFileHandler *os.File
	Err               error
}

func (shell Shell) executionPlanner(executionCtx context.Context) executionPlan {
	if shell.runtimeSettings.ShouldUseSubShell {
		subShellCmd := shell.runtimeSettings.Command + " " +
			strings.Join(shell.runtimeSettings.Args, " ")
		subShellArgs := []string{"-c", "source /etc/profile; " + subShellCmd}
		shell.runtimeSettings.Command = "bash"
		shell.runtimeSettings.Args = subShellArgs
	}

	execCmd := exec.CommandContext(
		executionCtx, shell.runtimeSettings.Command, shell.runtimeSettings.Args...,
	)
	execCmd.Cancel = func() error {
		return execCmd.Process.Signal(syscall.SIGTERM)
	}
	execCmd.WaitDelay = time.Duration(ShellExecutionTimeoutGraceSecs) * time.Second

	var targetUserPtr *user.User
	if shell.runtimeSettings.Username != "" {
		sysCallCredentials, targetUser, err := shell.sysCallCredentialsFactory()
		if err != nil && !shell.runtimeSettings.ShouldIgnoreUsernameLookupError {
			return executionPlan{Err: err}
		}
		if err == nil {
			targetUserPtr = targetUser
			execCmd.SysProcAttr = &syscall.SysProcAttr{Credential: sysCallCredentials}
		}
	}

	if shell.runtimeSettings.WorkingDirectory != "" {
		workingDirectory, absErr := filepath.Abs(shell.runtimeSettings.WorkingDirectory)
		if absErr != nil {
			return executionPlan{Err: absErr}
		}
		execCmd.Dir = workingDirectory
	}

	var stdoutBytesBuffer bytes.Buffer
	var stdoutFileHandlerPtr *os.File
	execCmd.Stdout = &stdoutBytesBuffer
	if shell.runtimeSettings.StdoutFilePath != "" {
		stdoutFileHandler, err := os.Create(shell.runtimeSettings.StdoutFilePath)
		if err != nil {
			return executionPlan{Err: err}
		}
		stdoutFileHandlerPtr = stdoutFileHandler
		execCmd.Stdout = stdoutFileHandlerPtr
	}

	var stderrBytesBuffer bytes.Buffer
	var stderrFileHandlerPtr *os.File
	execCmd.Stderr = &stderrBytesBuffer
	if shell.runtimeSettings.StderrFilePath != "" {
		stderrFileHandler, err := os.Create(shell.runtimeSettings.StderrFilePath)
		if err != nil {
			if closeErr := shell.closeFileHandler(stdoutFileHandlerPtr); closeErr != nil {
				slog.Error(
					"ShellStdoutFileCloseFailed",
					slog.String("file", stdoutFileHandlerPtr.Name()),
					slog.String("err", closeErr.Error()),
				)
			}
			return executionPlan{Err: err}
		}
		stderrFileHandlerPtr = stderrFileHandler
		execCmd.Stderr = stderrFileHandlerPtr
	}

	execCmd.Env = shell.childEnvironmentBuilder(execCmd, targetUserPtr)

	return executionPlan{
		ExecCmd:           execCmd,
		StdoutBytesBuffer: &stdoutBytesBuffer,
		StdoutFileHandler: stdoutFileHandlerPtr,
		StderrBytesBuffer: &stderrBytesBuffer,
		StderrFileHandler: stderrFileHandlerPtr,
	}
}

func (shell Shell) executionTimeoutResolver() time.Duration {
	timeoutSecs := shell.runtimeSettings.ExecutionTimeoutSecs
	if timeoutSecs == 0 {
		timeoutSecs = ShellExecutionTimeoutDefaultSecs
	}
	if timeoutSecs > ShellExecutionTimeoutHardLimitSecs &&
		!shell.runtimeSettings.ShouldDisableTimeoutHardLimit {
		timeoutSecs = ShellExecutionTimeoutHardLimitSecs
	}

	return time.Duration(timeoutSecs) * time.Second
}

func (shell Shell) Run() (stdoutStr string, err error) {
	executionCtx, cancelExecution := context.WithTimeout(
		context.Background(), shell.executionTimeoutResolver(),
	)
	defer cancelExecution()

	runPlan := shell.executionPlanner(executionCtx)
	if runPlan.Err != nil {
		return stdoutStr, runPlan.Err
	}

	runPlan.Err = runPlan.ExecCmd.Run()
	if closeErr := shell.closeFileHandler(runPlan.StdoutFileHandler); closeErr != nil {
		slog.Error(
			"ShellStdoutFileCloseFailed",
			slog.String("file", runPlan.StdoutFileHandler.Name()),
			slog.String("err", closeErr.Error()),
		)
	}
	if closeErr := shell.closeFileHandler(runPlan.StderrFileHandler); closeErr != nil {
		slog.Error(
			"ShellStderrFileCloseFailed",
			slog.String("file", runPlan.StderrFileHandler.Name()),
			slog.String("err", closeErr.Error()),
		)
	}

	if runPlan.StdoutBytesBuffer != nil {
		stdoutStr = strings.TrimSpace(runPlan.StdoutBytesBuffer.String())
	}
	if runPlan.Err == nil {
		return stdoutStr, nil
	}

	// exec.Wait prefers the child's own ExitError over the context error, so
	// cancellation is detected on the context — errors.Is(err, DeadlineExceeded)
	// never matches a signal-killed command.
	if errors.Is(executionCtx.Err(), context.DeadlineExceeded) {
		return stdoutStr, &ShellError{
			StdErr:   "CommandDeadlineExceeded",
			ExitCode: ShellCommandTimeoutExitCode,
		}
	}
	if exitErr, assertOk := runPlan.Err.(*exec.ExitError); assertOk {
		return stdoutStr, &ShellError{
			StdErr:   runPlan.StderrBytesBuffer.String(),
			ExitCode: exitErr.ExitCode(),
		}
	}

	return stdoutStr, runPlan.Err
}
