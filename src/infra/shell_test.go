package tkInfra

import (
	"context"
	"math"
	"os"
	"os/user"
	"path/filepath"
	"slices"
	"strconv"
	"testing"
	"time"

	tkValueObject "github.com/goinfinite/tk/src/domain/valueObject"
)

func TestShell(test *testing.T) {
	unresolvableUserIdFinder := func() uint32 {
		for candidate := uint32(999999); candidate > 0; candidate-- {
			_, lookupErr := user.LookupId(strconv.FormatUint(uint64(candidate), 10))
			if lookupErr != nil {
				return candidate
			}
		}

		test.Fatal("NoUnresolvableUserIdFound")
		return 0
	}
	unresolvableUserId := unresolvableUserIdFinder()

	test.Run("BasicShell", func(subtest *testing.T) {
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
				subtest.Errorf("MissingExpectedError: [%s %v]", testCase.command, testCase.args)
			}
			if !testCase.expectError && err != nil {
				subtest.Errorf("UnexpectedError: '%s' [%s %v]", err.Error(), testCase.command, testCase.args)
			}
			if !testCase.expectError && shellOutput != testCase.expectedOutput {
				subtest.Errorf(
					"UnexpectedOutputValue: '%s' vs '%s' [%s %v]",
					shellOutput, testCase.expectedOutput, testCase.command, testCase.args,
				)
			}
		}
	})

	test.Run("ErrorHandling", func(subtest *testing.T) {
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
				subtest.Errorf("MissingExpectedError: [%s %v]", testCase.command, testCase.args)
			}

			if err != nil {
				if shellErr, assertOk := err.(*ShellError); assertOk {
					if shellErr.ExitCode == 0 {
						subtest.Errorf("UnexpectedZeroExitCode: [%s %v]", testCase.command, testCase.args)
					}
				}
			}
		}
	})

	test.Run("StdoutFileCapture", func(subtest *testing.T) {
		outputFilePath := filepath.Join(subtest.TempDir(), "stdout.log")

		shell := NewShell(ShellSettings{
			Command:        "echo",
			Args:           []string{"hello"},
			StdoutFilePath: outputFilePath,
		})
		_, err := shell.Run()
		if err != nil {
			subtest.Fatalf("RunFailed: %v", err)
		}

		fileContent, readErr := os.ReadFile(outputFilePath)
		if readErr != nil {
			subtest.Fatalf("OutputFileReadFailed: %v", readErr)
		}
		if string(fileContent) != "hello\n" {
			subtest.Errorf("UnexpectedFileContent: '%s'", fileContent)
		}
	})

	test.Run("CaptureFileHandlesDoNotLeak", func(subtest *testing.T) {
		openDescriptorCount := func() int {
			fdEntries, err := os.ReadDir("/proc/self/fd")
			if err != nil {
				subtest.Fatalf("FdScanFailed: %v", err)
			}
			return len(fdEntries)
		}

		captureDir := subtest.TempDir()
		settings := ShellSettings{
			Command:        "echo",
			Args:           []string{"hello"},
			StdoutFilePath: filepath.Join(captureDir, "stdout.log"),
			StderrFilePath: filepath.Join(captureDir, "stderr.log"),
		}

		runCapture := func() {
			_, runErr := NewShell(settings).Run()
			if runErr != nil {
				subtest.Fatalf("RunFailed: %v", runErr)
			}
		}

		runCapture()
		baselineDescriptors := openDescriptorCount()

		for range 9 {
			runCapture()
		}

		if leakedCount := openDescriptorCount() - baselineDescriptors; leakedCount > 0 {
			subtest.Errorf("CaptureFileHandlesLeaked: %d descriptors left open", leakedCount)
		}
	})

	test.Run("CommandTimeoutEnforced", func(subtest *testing.T) {
		_, err := NewShell(ShellSettings{
			Command:              "sleep",
			Args:                 []string{"5"},
			ExecutionTimeoutSecs: 1,
		}).Run()

		shellErr, assertOk := err.(*ShellError)
		if !assertOk {
			subtest.Fatalf("ExpectedShellError,Got%v", err)
		}
		if shellErr.ExitCode != ShellCommandTimeoutExitCode {
			subtest.Errorf("UnexpectedTimeoutExitCode: %d", shellErr.ExitCode)
		}
		if shellErr.StdErr != "CommandDeadlineExceeded" {
			subtest.Errorf("UnexpectedTimeoutMessage: '%s'", shellErr.StdErr)
		}
	})

	test.Run("NaturalExitCode124KeepsCommandStdErr", func(subtest *testing.T) {
		_, err := NewShell(ShellSettings{
			Command: "bash",
			Args:    []string{"-c", "echo boom 1>&2; exit 124"},
		}).Run()

		shellErr, assertOk := err.(*ShellError)
		if !assertOk {
			subtest.Fatalf("ExpectedShellError,Got%v", err)
		}
		if shellErr.ExitCode != 124 {
			subtest.Errorf("UnexpectedExitCode: %d", shellErr.ExitCode)
		}
		if shellErr.StdErr != "boom\n" {
			subtest.Errorf("Natural124MisreportedAsTimeout: '%s'", shellErr.StdErr)
		}
	})

	test.Run("ShouldDisableTimeoutLetsLongCommandFinish", func(subtest *testing.T) {
		_, err := NewShell(ShellSettings{
			Command:              "sleep",
			Args:                 []string{"2"},
			ExecutionTimeoutSecs: 1,
			ShouldDisableTimeout: true,
		}).Run()
		if err != nil {
			subtest.Errorf("CommandKilledDespiteDisabledTimeout: %v", err)
		}
	})

	test.Run("ExecutionDeadlineEnforced", func(subtest *testing.T) {
		deadline := tkValueObject.NewUnixTimeAfterNow(1 * time.Second)

		_, err := NewShell(ShellSettings{
			Command:           "sleep",
			Args:              []string{"5"},
			ExecutionDeadline: &deadline,
		}).Run()

		shellErr, assertOk := err.(*ShellError)
		if !assertOk {
			subtest.Fatalf("ExpectedShellError,Got%v", err)
		}
		if shellErr.ExitCode != ShellCommandTimeoutExitCode {
			subtest.Errorf("UnexpectedTimeoutExitCode: %d", shellErr.ExitCode)
		}
	})

	test.Run("PastExecutionDeadlineTimesOutImmediately", func(subtest *testing.T) {
		pastDeadline := tkValueObject.NewUnixTimeBeforeNow(1 * time.Hour)

		_, err := NewShell(ShellSettings{
			Command:           "sleep",
			Args:              []string{"5"},
			ExecutionDeadline: &pastDeadline,
		}).Run()

		shellErr, assertOk := err.(*ShellError)
		if !assertOk {
			subtest.Fatalf("ExpectedShellError,Got%v", err)
		}
		if shellErr.ExitCode != ShellCommandTimeoutExitCode {
			subtest.Errorf("UnexpectedTimeoutExitCode: %d", shellErr.ExitCode)
		}
	})

	test.Run("ShouldDisableTimeoutOverridesDeadline", func(subtest *testing.T) {
		pastDeadline := tkValueObject.NewUnixTimeBeforeNow(1 * time.Hour)

		_, err := NewShell(ShellSettings{
			Command:              "echo",
			ShouldDisableTimeout: true,
			ExecutionDeadline:    &pastDeadline,
		}).Run()

		if err != nil {
			subtest.Errorf("ExpectedSuccess,Got%v", err)
		}
	})

	test.Run("UserIdRunsCommandAsTargetAccount", func(subtest *testing.T) {
		if os.Geteuid() != 0 {
			subtest.Skip("RootPrivilegesRequired")
		}

		nobody, lookupErr := user.Lookup("nobody")
		if lookupErr != nil {
			subtest.Skipf("NobodyUserMissing: %v", lookupErr)
		}
		nobodyUid, uidErr := strconv.Atoi(nobody.Uid)
		if uidErr != nil {
			subtest.Fatalf("UidParseFailed: %v", uidErr)
		}
		nobodyUserId := uint32(nobodyUid)

		uidStr, err := NewShell(ShellSettings{
			Command: "id",
			Args:    []string{"-u"},
			UserId:  &nobodyUserId,
		}).Run()
		if err != nil {
			subtest.Fatalf("RunFailed: %v", err)
		}
		if uidStr != nobody.Uid {
			subtest.Errorf("UidMismatch: %s vs %s", uidStr, nobody.Uid)
		}

		gidStr, err := NewShell(ShellSettings{
			Command: "id",
			Args:    []string{"-g"},
			UserId:  &nobodyUserId,
		}).Run()
		if err != nil {
			subtest.Fatalf("RunFailed: %v", err)
		}
		if gidStr != nobody.Gid {
			subtest.Errorf("GidMismatch: %s vs %s", gidStr, nobody.Gid)
		}
	})

	test.Run("UserIdZeroRequestsRootAccount", func(subtest *testing.T) {
		rootUserId := uint32(0)

		runPlan := NewShell(ShellSettings{
			Command: "true",
			UserId:  &rootUserId,
		}).executionPlanner(context.Background())
		if runPlan.Err != nil {
			subtest.Fatalf("ExecutionPlanningFailed: %v", runPlan.Err)
		}

		sysProcAttr := runPlan.ExecCmd.SysProcAttr
		if sysProcAttr == nil || sysProcAttr.Credential == nil {
			subtest.Fatalf("MissingSysCallCredentialsForRootUserId")
		}
		if sysProcAttr.Credential.Uid != 0 {
			subtest.Errorf("UnexpectedCredentialUid: %d", sysProcAttr.Credential.Uid)
		}
	})

	test.Run("UnresolvableUserIdFailsLoud", func(subtest *testing.T) {
		_, err := NewShell(ShellSettings{
			Command: "true",
			UserId:  &unresolvableUserId,
		}).Run()
		if err == nil {
			subtest.Errorf("MissingErrorForUnresolvableUserId")
		}
	})

	test.Run("UnresolvableUserIdIgnoredByFlag", func(subtest *testing.T) {
		stdoutStr, err := NewShell(ShellSettings{
			Command:                         "id",
			Args:                            []string{"-u"},
			UserId:                          &unresolvableUserId,
			ShouldIgnoreUsernameLookupError: true,
		}).Run()
		if err != nil {
			subtest.Fatalf("RunFailed: %v", err)
		}
		if stdoutStr != strconv.Itoa(os.Getuid()) {
			subtest.Errorf("ExpectedCurrentUser: %s vs %d", stdoutStr, os.Getuid())
		}
	})

	test.Run("UsernameWinsOverUserId", func(subtest *testing.T) {
		if os.Geteuid() != 0 {
			subtest.Skip("RootPrivilegesRequired")
		}

		stdoutStr, err := NewShell(ShellSettings{
			Command:  "id",
			Args:     []string{"-u"},
			Username: "root",
			UserId:   &unresolvableUserId,
		}).Run()
		if err != nil {
			subtest.Fatalf("RunFailed: %v", err)
		}
		if stdoutStr != "0" {
			subtest.Errorf("ExpectedRootUid: %s", stdoutStr)
		}
	})
}

func TestShellExecutionTimeoutResolver(test *testing.T) {
	secondsToDuration := func(seconds uint64) time.Duration {
		return time.Duration(seconds) * time.Second
	}

	test.Run("DefaultWhenUnset", func(subtest *testing.T) {
		timeout := NewShell(ShellSettings{}).executionTimeoutResolver()
		expectedTimeout := secondsToDuration(ShellExecutionTimeoutDefaultSecs)

		if timeout != expectedTimeout {
			subtest.Errorf("UnexpectedTimeout: %s vs %s", timeout, expectedTimeout)
		}
	})

	test.Run("RequestedTimeoutAboveLegacyCapIsHonored", func(subtest *testing.T) {
		timeout := NewShell(ShellSettings{
			ExecutionTimeoutSecs: 7200,
		}).executionTimeoutResolver()
		expectedTimeout := secondsToDuration(7200)

		if timeout != expectedTimeout {
			subtest.Errorf("UnexpectedTimeout: %s vs %s", timeout, expectedTimeout)
		}
	})

	test.Run("HugeRequestedTimeoutSaturatesWithoutOverflow", func(subtest *testing.T) {
		timeout := NewShell(ShellSettings{
			ExecutionTimeoutSecs: math.MaxUint64,
		}).executionTimeoutResolver()

		if timeout <= 0 {
			subtest.Errorf("HugeTimeoutOverflowed: %s", timeout)
		}
	})

	test.Run("EarlierDeadlineWinsOverDuration", func(subtest *testing.T) {
		deadline := tkValueObject.NewUnixTimeAfterNow(120 * time.Second)

		timeout := NewShell(ShellSettings{
			ExecutionTimeoutSecs: 600,
			ExecutionDeadline:    &deadline,
		}).executionTimeoutResolver()

		if timeout > 121*time.Second || timeout < 118*time.Second {
			subtest.Errorf("DeadlineDidNotWin: %s", timeout)
		}
	})

	test.Run("EarlierDurationWinsOverDeadline", func(subtest *testing.T) {
		deadline := tkValueObject.NewUnixTimeAfterNow(6000 * time.Second)

		timeout := NewShell(ShellSettings{
			ExecutionTimeoutSecs: 600,
			ExecutionDeadline:    &deadline,
		}).executionTimeoutResolver()
		expectedTimeout := secondsToDuration(600)

		if timeout != expectedTimeout {
			subtest.Errorf("UnexpectedTimeout: %s vs %s", timeout, expectedTimeout)
		}
	})

	test.Run("DeadlineAloneDefinesTimeout", func(subtest *testing.T) {
		deadline := tkValueObject.NewUnixTimeAfterNow(3000 * time.Second)

		timeout := NewShell(ShellSettings{
			ExecutionDeadline: &deadline,
		}).executionTimeoutResolver()

		if timeout > 3001*time.Second || timeout < 2998*time.Second {
			subtest.Errorf("DeadlineAloneNotHonored: %s", timeout)
		}
	})

	test.Run("FarDeadlineAloneIsHonored", func(subtest *testing.T) {
		deadline := tkValueObject.NewUnixTimeAfterNow(7200 * time.Second)

		timeout := NewShell(ShellSettings{
			ExecutionDeadline: &deadline,
		}).executionTimeoutResolver()

		if timeout > 7201*time.Second || timeout < 7198*time.Second {
			subtest.Errorf("FarDeadlineNotHonored: %s", timeout)
		}
	})
}

func TestChildEnvironment(test *testing.T) {
	test.Setenv("TK_PARENT_ONLY_VAR", "leaked")
	standardSystemPath := "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

	test.Run("InheritsParentEnvAndWorkingDirectoryPwd", func(subtest *testing.T) {
		runPlan := NewShell(ShellSettings{
			Command:          "env",
			WorkingDirectory: "/tmp",
		}).executionPlanner(context.Background())
		if runPlan.Err != nil {
			subtest.Fatalf("ExecutionPlanningFailed: %v", runPlan.Err)
		}
		childEnv := runPlan.ExecCmd.Env

		if !slices.Contains(childEnv, "TK_PARENT_ONLY_VAR=leaked") {
			subtest.Errorf("ParentEnvNotInherited: %v", childEnv)
		}
		if !slices.Contains(childEnv, "DEBIAN_FRONTEND=noninteractive") {
			subtest.Errorf("MissingDebianFrontend: %v", childEnv)
		}
		if !slices.Contains(childEnv, "PWD=/tmp") {
			subtest.Errorf("WorkingDirectoryPwdNotPropagated: %v", childEnv)
		}
	})

	test.Run("UsernameAloneStillInheritsParentEnv", func(subtest *testing.T) {
		runPlan := NewShell(ShellSettings{
			Command:  "env",
			Username: "root",
		}).executionPlanner(context.Background())
		if runPlan.Err != nil {
			subtest.Fatalf("ExecutionPlanningFailed: %v", runPlan.Err)
		}
		childEnv := runPlan.ExecCmd.Env

		if !slices.Contains(childEnv, "TK_PARENT_ONLY_VAR=leaked") {
			subtest.Errorf("UsernameShouldNotImplyCleanEnv: %v", childEnv)
		}
	})

	test.Run("CleanEnvDropsParentWhenFlagSet", func(subtest *testing.T) {
		runPlan := NewShell(ShellSettings{
			Command:           "env",
			ShouldUseCleanEnv: true,
			WorkingDirectory:  "/tmp",
			Envs:              []string{"HOME=/home/target"},
		}).executionPlanner(context.Background())
		if runPlan.Err != nil {
			subtest.Fatalf("ExecutionPlanningFailed: %v", runPlan.Err)
		}
		childEnv := runPlan.ExecCmd.Env

		if slices.Contains(childEnv, "TK_PARENT_ONLY_VAR=leaked") {
			subtest.Errorf("ParentEnvLeakedIntoCleanChild: %v", childEnv)
		}
		if !slices.Contains(childEnv, "DEBIAN_FRONTEND=noninteractive") {
			subtest.Errorf("MissingDebianFrontend: %v", childEnv)
		}
		if !slices.Contains(childEnv, "PWD=/tmp") {
			subtest.Errorf("WorkingDirectoryPwdNotPropagated: %v", childEnv)
		}
		if !slices.Contains(childEnv, "HOME=/home/target") {
			subtest.Errorf("MissingExplicitEnv: %v", childEnv)
		}
		expectedPath := "PATH=" + os.Getenv("HOME") + "/.local/bin:" + standardSystemPath
		if !slices.Contains(childEnv, expectedPath) {
			subtest.Errorf("MissingUserLocalBinAndStandardPath: %v", childEnv)
		}
	})

	test.Run("CleanEnvKeepsParentHomeForSameUser", func(subtest *testing.T) {
		subtest.Setenv("HOME", "/home/simulator-user")

		runPlan := NewShell(ShellSettings{
			Command:           "env",
			ShouldUseCleanEnv: true,
		}).executionPlanner(context.Background())
		if runPlan.Err != nil {
			subtest.Fatalf("ExecutionPlanningFailed: %v", runPlan.Err)
		}
		childEnv := runPlan.ExecCmd.Env

		if !slices.Contains(childEnv, "HOME=/home/simulator-user") {
			subtest.Errorf("ParentHomeDroppedForSameUser: %v", childEnv)
		}
	})

	test.Run("CleanEnvSetsTargetUserHomeOnUsernameSwitch", func(subtest *testing.T) {
		targetUser, err := user.Lookup("root")
		if err != nil {
			subtest.Fatalf("RootLookupFailed: %v", err)
		}

		runPlan := NewShell(ShellSettings{
			Command:           "env",
			ShouldUseCleanEnv: true,
			Username:          "root",
		}).executionPlanner(context.Background())
		if runPlan.Err != nil {
			subtest.Fatalf("ExecutionPlanningFailed: %v", runPlan.Err)
		}
		childEnv := runPlan.ExecCmd.Env

		if !slices.Contains(childEnv, "HOME="+targetUser.HomeDir) {
			subtest.Errorf("TargetUserHomeMissing: %v", childEnv)
		}
		if parentHome := os.Getenv("HOME"); parentHome != targetUser.HomeDir {
			if slices.Contains(childEnv, "HOME="+parentHome) {
				subtest.Errorf("ParentHomeLeakedIntoSwitchedChild: %v", childEnv)
			}
			targetLocalBinPath := "PATH=" + targetUser.HomeDir + "/.local/bin:" +
				standardSystemPath
			if !slices.Contains(childEnv, targetLocalBinPath) {
				subtest.Errorf("TargetUserLocalBinPathMissing: %v", childEnv)
			}
		}
	})

	test.Run("CleanEnvPwdIsAbsoluteForRelativeWorkingDirectory", func(subtest *testing.T) {
		runPlan := NewShell(ShellSettings{
			Command:           "env",
			ShouldUseCleanEnv: true,
			WorkingDirectory:  "relative-dir",
		}).executionPlanner(context.Background())
		if runPlan.Err != nil {
			subtest.Fatalf("ExecutionPlanningFailed: %v", runPlan.Err)
		}

		expectedPwd, err := filepath.Abs("relative-dir")
		if err != nil {
			subtest.Fatalf("AbsPathFailed: %v", err)
		}
		if !slices.Contains(runPlan.ExecCmd.Env, "PWD="+expectedPwd) {
			subtest.Errorf("RelativeWorkingDirectoryPwdNotAbsolute: %v", runPlan.ExecCmd.Env)
		}
	})

	test.Run("EnvsWinOverSameNamedParentEntries", func(subtest *testing.T) {
		subtest.Setenv("TK_TEST_ENV_PRECEDENCE", "from-parent")

		shellOutput, err := NewShell(ShellSettings{
			Command: "printenv",
			Args:    []string{"TK_TEST_ENV_PRECEDENCE"},
			Envs:    []string{"TK_TEST_ENV_PRECEDENCE=from-envs"},
		}).Run()
		if err != nil {
			subtest.Fatalf("RunFailed: %v", err)
		}
		if shellOutput != "from-envs" {
			subtest.Errorf("EnvsDidNotOverrideParentEntry: got '%s'", shellOutput)
		}
	})

	test.Run("EnvsWinOverCleanBaseEntries", func(subtest *testing.T) {
		subtest.Setenv("HOME", "/parent/home")

		shellOutput, err := NewShell(ShellSettings{
			Command:           "printenv",
			Args:              []string{"HOME"},
			ShouldUseCleanEnv: true,
			Envs:              []string{"HOME=/home/envs-wins"},
		}).Run()
		if err != nil {
			subtest.Fatalf("RunFailed: %v", err)
		}
		if shellOutput != "/home/envs-wins" {
			subtest.Errorf("EnvsDidNotOverrideCleanBaseHome: got '%s'", shellOutput)
		}
	})
}

func TestIsStdoutTerminal(test *testing.T) {
	test.Run("PipeIsNotTerminal", func(subtest *testing.T) {
		originalStdout := os.Stdout

		readEnd, writeEnd, err := os.Pipe()
		if err != nil {
			subtest.Fatalf("PipeCreationFailed: %v", err)
		}
		defer func() { _ = writeEnd.Close() }()
		defer func() { _ = readEnd.Close() }()
		defer func() { os.Stdout = originalStdout }()

		os.Stdout = readEnd
		if IsStdoutTerminal() {
			subtest.Errorf("PipeReportedAsTerminal")
		}
	})
}
