package tkInfra

import (
	"context"
	"os"
	"os/user"
	"path/filepath"
	"slices"
	"strconv"
	"testing"
)

func TestShell(t *testing.T) {
	t.Run("BasicShell", func(t *testing.T) {
		testCaseStructs := []struct {
			command        string
			args           []string
			expectedOutput string
			expectError    bool
		}{
			{"echo", []string{"hello"}, "hello", false},
			{"echo", []string{"-n", "test"}, "test", false},
			{"true", []string{}, "", false},
			{"false", []string{}, "", true},
		}

		for _, testCase := range testCaseStructs {
			shell := NewShell(
				ShellSettings{Command: testCase.command, Args: testCase.args},
			)
			shellOutput, err := shell.Run()
			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%s %v]", testCase.command, testCase.args)
			}
			if !testCase.expectError && err != nil {
				t.Errorf("UnexpectedError: '%s' [%s %v]", err.Error(), testCase.command, testCase.args)
			}
			if !testCase.expectError && shellOutput != testCase.expectedOutput {
				t.Errorf(
					"UnexpectedOutputValue: '%s' vs '%s' [%s %v]",
					shellOutput, testCase.expectedOutput, testCase.command, testCase.args,
				)
			}
		}
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		testCaseStructs := []struct {
			command     string
			args        []string
			expectError bool
		}{
			{"nonexistentcommand", []string{}, true},
			{"ls", []string{"/nonexistent/path"}, true},
			{"cat", []string{"/nonexistent/file"}, true},
		}

		for _, testCase := range testCaseStructs {
			shell := NewShell(
				ShellSettings{Command: testCase.command, Args: testCase.args},
			)

			_, err := shell.Run()
			if testCase.expectError && err == nil {
				t.Errorf("MissingExpectedError: [%s %v]", testCase.command, testCase.args)
			}

			if err != nil {
				if shellErr, assertOk := err.(*ShellError); assertOk {
					if shellErr.ExitCode == 0 {
						t.Errorf("UnexpectedZeroExitCode: [%s %v]", testCase.command, testCase.args)
					}
				}
			}
		}
	})

	t.Run("StdoutFileCapture", func(t *testing.T) {
		outputFilePath := filepath.Join(t.TempDir(), "stdout.log")

		shell := NewShell(ShellSettings{
			Command:        "echo",
			Args:           []string{"hello"},
			StdoutFilePath: outputFilePath,
		})
		_, err := shell.Run()
		if err != nil {
			t.Fatalf("RunFailed: %v", err)
		}

		fileContent, readErr := os.ReadFile(outputFilePath)
		if readErr != nil {
			t.Fatalf("OutputFileReadFailed: %v", readErr)
		}
		if string(fileContent) != "hello\n" {
			t.Errorf("UnexpectedFileContent: '%s'", fileContent)
		}
	})

	t.Run("CaptureFileHandlesDoNotLeak", func(t *testing.T) {
		openDescriptorCount := func() int {
			fdEntries, err := os.ReadDir("/proc/self/fd")
			if err != nil {
				t.Fatalf("FdScanFailed: %v", err)
			}
			return len(fdEntries)
		}

		captureDir := t.TempDir()
		settings := ShellSettings{
			Command:        "echo",
			Args:           []string{"hello"},
			StdoutFilePath: filepath.Join(captureDir, "stdout.log"),
			StderrFilePath: filepath.Join(captureDir, "stderr.log"),
		}

		runCapture := func() {
			if _, runErr := NewShell(settings).Run(); runErr != nil {
				t.Fatalf("RunFailed: %v", runErr)
			}
		}

		runCapture()
		baselineDescriptors := openDescriptorCount()

		for range 9 {
			runCapture()
		}

		if leakedCount := openDescriptorCount() - baselineDescriptors; leakedCount > 0 {
			t.Errorf("CaptureFileHandlesLeaked: %d descriptors left open", leakedCount)
		}
	})

	t.Run("CommandTimeoutEnforced", func(t *testing.T) {
		_, err := NewShell(ShellSettings{
			Command:              "sleep",
			Args:                 []string{"5"},
			ExecutionTimeoutSecs: 1,
		}).Run()

		shellErr, assertOk := err.(*ShellError)
		if !assertOk {
			t.Fatalf("ExpectedShellError,Got%v", err)
		}
		if shellErr.ExitCode != ShellCommandTimeoutExitCode {
			t.Errorf("UnexpectedTimeoutExitCode: %d", shellErr.ExitCode)
		}
		if shellErr.StdErr != "CommandDeadlineExceeded" {
			t.Errorf("UnexpectedTimeoutMessage: '%s'", shellErr.StdErr)
		}
	})

	t.Run("NaturalExitCode124KeepsCommandStdErr", func(t *testing.T) {
		_, err := NewShell(ShellSettings{
			Command: "bash",
			Args:    []string{"-c", "echo boom 1>&2; exit 124"},
		}).Run()

		shellErr, assertOk := err.(*ShellError)
		if !assertOk {
			t.Fatalf("ExpectedShellError,Got%v", err)
		}
		if shellErr.ExitCode != 124 {
			t.Errorf("UnexpectedExitCode: %d", shellErr.ExitCode)
		}
		if shellErr.StdErr != "boom\n" {
			t.Errorf("Natural124MisreportedAsTimeout: '%s'", shellErr.StdErr)
		}
	})

	t.Run("ShouldDisableTimeoutLetsLongCommandFinish", func(t *testing.T) {
		_, err := NewShell(ShellSettings{
			Command:              "sleep",
			Args:                 []string{"2"},
			ExecutionTimeoutSecs: 1,
			ShouldDisableTimeout: true,
		}).Run()
		if err != nil {
			t.Errorf("CommandKilledDespiteDisabledTimeout: %v", err)
		}
	})

	t.Run("UserIdRunsCommandAsTargetAccount", func(t *testing.T) {
		if os.Geteuid() != 0 {
			t.Skip("RootPrivilegesRequired")
		}

		nobody, lookupErr := user.Lookup("nobody")
		if lookupErr != nil {
			t.Skipf("NobodyUserMissing: %v", lookupErr)
		}
		nobodyUid, uidErr := strconv.Atoi(nobody.Uid)
		if uidErr != nil {
			t.Fatalf("UidParseFailed: %v", uidErr)
		}

		uidStr, err := NewShell(ShellSettings{
			Command: "id",
			Args:    []string{"-u"},
			UserId:  uint32(nobodyUid),
		}).Run()
		if err != nil {
			t.Fatalf("RunFailed: %v", err)
		}
		if uidStr != nobody.Uid {
			t.Errorf("UidMismatch: %s vs %s", uidStr, nobody.Uid)
		}

		gidStr, err := NewShell(ShellSettings{
			Command: "id",
			Args:    []string{"-g"},
			UserId:  uint32(nobodyUid),
		}).Run()
		if err != nil {
			t.Fatalf("RunFailed: %v", err)
		}
		if gidStr != nobody.Gid {
			t.Errorf("GidMismatch: %s vs %s", gidStr, nobody.Gid)
		}
	})

	t.Run("UnresolvableUserIdFailsLoud", func(t *testing.T) {
		_, err := NewShell(ShellSettings{
			Command: "true",
			UserId:  999999,
		}).Run()
		if err == nil {
			t.Errorf("MissingErrorForUnresolvableUserId")
		}
	})

	t.Run("UnresolvableUserIdIgnoredByFlag", func(t *testing.T) {
		stdoutStr, err := NewShell(ShellSettings{
			Command:                         "id",
			Args:                            []string{"-u"},
			UserId:                          999999,
			ShouldIgnoreUsernameLookupError: true,
		}).Run()
		if err != nil {
			t.Fatalf("RunFailed: %v", err)
		}
		if stdoutStr != strconv.Itoa(os.Getuid()) {
			t.Errorf("ExpectedCurrentUser: %s vs %d", stdoutStr, os.Getuid())
		}
	})

	t.Run("UsernameWinsOverUserId", func(t *testing.T) {
		if os.Geteuid() != 0 {
			t.Skip("RootPrivilegesRequired")
		}

		stdoutStr, err := NewShell(ShellSettings{
			Command:  "id",
			Args:     []string{"-u"},
			Username: "root",
			UserId:   999999,
		}).Run()
		if err != nil {
			t.Fatalf("RunFailed: %v", err)
		}
		if stdoutStr != "0" {
			t.Errorf("ExpectedRootUid: %s", stdoutStr)
		}
	})
}

func TestChildEnvironment(t *testing.T) {
	t.Setenv("TK_PARENT_ONLY_VAR", "leaked")
	standardSystemPath := "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

	t.Run("InheritsParentEnvAndWorkingDirectoryPwd", func(t *testing.T) {
		runPlan := NewShell(ShellSettings{
			Command:          "env",
			WorkingDirectory: "/tmp",
		}).executionPlanner(context.Background())
		if runPlan.Err != nil {
			t.Fatalf("ExecutionPlanningFailed: %v", runPlan.Err)
		}
		childEnv := runPlan.ExecCmd.Env

		if !slices.Contains(childEnv, "TK_PARENT_ONLY_VAR=leaked") {
			t.Errorf("ParentEnvNotInherited: %v", childEnv)
		}
		if !slices.Contains(childEnv, "DEBIAN_FRONTEND=noninteractive") {
			t.Errorf("MissingDebianFrontend: %v", childEnv)
		}
		if !slices.Contains(childEnv, "PWD=/tmp") {
			t.Errorf("WorkingDirectoryPwdNotPropagated: %v", childEnv)
		}
	})

	t.Run("UsernameAloneStillInheritsParentEnv", func(t *testing.T) {
		runPlan := NewShell(ShellSettings{
			Command:  "env",
			Username: "root",
		}).executionPlanner(context.Background())
		if runPlan.Err != nil {
			t.Fatalf("ExecutionPlanningFailed: %v", runPlan.Err)
		}
		childEnv := runPlan.ExecCmd.Env

		if !slices.Contains(childEnv, "TK_PARENT_ONLY_VAR=leaked") {
			t.Errorf("UsernameShouldNotImplyCleanEnv: %v", childEnv)
		}
	})

	t.Run("CleanEnvDropsParentWhenFlagSet", func(t *testing.T) {
		runPlan := NewShell(ShellSettings{
			Command:           "env",
			ShouldUseCleanEnv: true,
			WorkingDirectory:  "/tmp",
			Envs:              []string{"HOME=/home/target"},
		}).executionPlanner(context.Background())
		if runPlan.Err != nil {
			t.Fatalf("ExecutionPlanningFailed: %v", runPlan.Err)
		}
		childEnv := runPlan.ExecCmd.Env

		if slices.Contains(childEnv, "TK_PARENT_ONLY_VAR=leaked") {
			t.Errorf("ParentEnvLeakedIntoCleanChild: %v", childEnv)
		}
		if !slices.Contains(childEnv, "DEBIAN_FRONTEND=noninteractive") {
			t.Errorf("MissingDebianFrontend: %v", childEnv)
		}
		if !slices.Contains(childEnv, "PWD=/tmp") {
			t.Errorf("WorkingDirectoryPwdNotPropagated: %v", childEnv)
		}
		if !slices.Contains(childEnv, "HOME=/home/target") {
			t.Errorf("MissingExplicitEnv: %v", childEnv)
		}
		expectedPath := "PATH=" + os.Getenv("HOME") + "/.local/bin:" + standardSystemPath
		if !slices.Contains(childEnv, expectedPath) {
			t.Errorf("MissingUserLocalBinAndStandardPath: %v", childEnv)
		}
	})

	t.Run("CleanEnvKeepsParentHomeForSameUser", func(t *testing.T) {
		t.Setenv("HOME", "/home/simulator-user")

		runPlan := NewShell(ShellSettings{
			Command:           "env",
			ShouldUseCleanEnv: true,
		}).executionPlanner(context.Background())
		if runPlan.Err != nil {
			t.Fatalf("ExecutionPlanningFailed: %v", runPlan.Err)
		}
		childEnv := runPlan.ExecCmd.Env

		if !slices.Contains(childEnv, "HOME=/home/simulator-user") {
			t.Errorf("ParentHomeDroppedForSameUser: %v", childEnv)
		}
	})

	t.Run("CleanEnvSetsTargetUserHomeOnUsernameSwitch", func(t *testing.T) {
		targetUser, err := user.Lookup("root")
		if err != nil {
			t.Fatalf("RootLookupFailed: %v", err)
		}

		runPlan := NewShell(ShellSettings{
			Command:           "env",
			ShouldUseCleanEnv: true,
			Username:          "root",
		}).executionPlanner(context.Background())
		if runPlan.Err != nil {
			t.Fatalf("ExecutionPlanningFailed: %v", runPlan.Err)
		}
		childEnv := runPlan.ExecCmd.Env

		if !slices.Contains(childEnv, "HOME="+targetUser.HomeDir) {
			t.Errorf("TargetUserHomeMissing: %v", childEnv)
		}
		if parentHome := os.Getenv("HOME"); parentHome != targetUser.HomeDir {
			if slices.Contains(childEnv, "HOME="+parentHome) {
				t.Errorf("ParentHomeLeakedIntoSwitchedChild: %v", childEnv)
			}
			targetLocalBinPath := "PATH=" + targetUser.HomeDir + "/.local/bin:" +
				standardSystemPath
			if !slices.Contains(childEnv, targetLocalBinPath) {
				t.Errorf("TargetUserLocalBinPathMissing: %v", childEnv)
			}
		}
	})

	t.Run("CleanEnvPwdIsAbsoluteForRelativeWorkingDirectory", func(t *testing.T) {
		runPlan := NewShell(ShellSettings{
			Command:           "env",
			ShouldUseCleanEnv: true,
			WorkingDirectory:  "relative-dir",
		}).executionPlanner(context.Background())
		if runPlan.Err != nil {
			t.Fatalf("ExecutionPlanningFailed: %v", runPlan.Err)
		}

		expectedPwd, err := filepath.Abs("relative-dir")
		if err != nil {
			t.Fatalf("AbsPathFailed: %v", err)
		}
		if !slices.Contains(runPlan.ExecCmd.Env, "PWD="+expectedPwd) {
			t.Errorf("RelativeWorkingDirectoryPwdNotAbsolute: %v", runPlan.ExecCmd.Env)
		}
	})

	t.Run("EnvsWinOverSameNamedParentEntries", func(t *testing.T) {
		t.Setenv("TK_TEST_ENV_PRECEDENCE", "from-parent")

		shellOutput, err := NewShell(ShellSettings{
			Command: "printenv",
			Args:    []string{"TK_TEST_ENV_PRECEDENCE"},
			Envs:    []string{"TK_TEST_ENV_PRECEDENCE=from-envs"},
		}).Run()
		if err != nil {
			t.Fatalf("RunFailed: %v", err)
		}
		if shellOutput != "from-envs" {
			t.Errorf("EnvsDidNotOverrideParentEntry: got '%s'", shellOutput)
		}
	})

	t.Run("EnvsWinOverCleanBaseEntries", func(t *testing.T) {
		t.Setenv("HOME", "/parent/home")

		shellOutput, err := NewShell(ShellSettings{
			Command:           "printenv",
			Args:              []string{"HOME"},
			ShouldUseCleanEnv: true,
			Envs:              []string{"HOME=/home/envs-wins"},
		}).Run()
		if err != nil {
			t.Fatalf("RunFailed: %v", err)
		}
		if shellOutput != "/home/envs-wins" {
			t.Errorf("EnvsDidNotOverrideCleanBaseHome: got '%s'", shellOutput)
		}
	})
}

func TestIsStdoutTerminal(t *testing.T) {
	t.Run("PipeIsNotTerminal", func(t *testing.T) {
		originalStdout := os.Stdout

		readEnd, writeEnd, err := os.Pipe()
		if err != nil {
			t.Fatalf("PipeCreationFailed: %v", err)
		}
		defer func() { _ = writeEnd.Close() }()
		defer func() { _ = readEnd.Close() }()
		defer func() { os.Stdout = originalStdout }()

		os.Stdout = readEnd
		if IsStdoutTerminal() {
			t.Errorf("PipeReportedAsTerminal")
		}
	})
}
