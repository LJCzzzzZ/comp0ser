package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

type Runner struct {
	Timeout time.Duration
}

type RunError struct {
	Command string
	Stdout  string
	Stderr  string
	Err     error
}

func (e *RunError) Error() string {
	msg := fmt.Sprintf("run failed: %v\ncmd: %s", e.Err, e.Command)
	if e.Stderr != "" {
		msg += "\nstderr:\n" + e.Stderr
	}
	if e.Stdout != "" {
		msg += "\nstdout:\n" + e.Stdout
	}
	return msg
}

func (r *Runner) Run(parent context.Context, cmd *Cmd) error {
	if cmd == nil {
		return errors.New("nil cmd")
	}

	if cmd.Bin == "" {
		return errors.New("no ffmpeg")
	}

	ctx := parent
	cancel := func() {}
	if r.Timeout > 0 {
		ctx, cancel = context.WithTimeout(parent, r.Timeout)
	}
	defer cancel()

	command := exec.CommandContext(ctx, cmd.Bin, cmd.Args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	command.Stdout = &stdoutBuf
	command.Stderr = &stderrBuf

	err := command.Run()

	// 如果是超时/取消，返回更清晰的错误信息
	if ctx.Err() != nil {
		return &RunError{
			Command: command.String(),
			Stdout:  stdoutBuf.String(),
			Stderr:  stderrBuf.String(),
			Err:     ctx.Err(),
		}
	}

	if err != nil {
		return &RunError{
			Command: command.String(),
			Stdout:  stdoutBuf.String(),
			Stderr:  stderrBuf.String(),
			Err:     err,
		}
	}

	return nil
}
