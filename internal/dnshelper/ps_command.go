package dnshelper

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/nrkno/terraform-provider-windns/internal/config"
	"golang.org/x/crypto/ssh"

	"github.com/masterzen/winrm"
)

type CreatePSCommandOpts struct {
	ForceArray bool
	JSONOutput bool
	JSONDepth  int
	Password   string
	Server     string
	Username   string
}

type PSCommand struct {
	CreatePSCommandOpts
	cmd string
}

func NewPSCommand(cmds []string, opts CreatePSCommandOpts) *PSCommand {
	cmd := strings.Join(cmds, " ")

	if opts.Server != "" {
		cmd = fmt.Sprintf("%s -ComputerName %s", cmd, opts.Server)
	}

	if opts.JSONOutput {
		cmd = fmt.Sprintf("%s %s", cmd, "| ConvertTo-Json")
		if opts.JSONDepth != 0 {
			cmd = fmt.Sprintf("%s %s", cmd, fmt.Sprintf("-Depth %d", opts.JSONDepth))
		}
	}

	res := PSCommand{
		CreatePSCommandOpts: opts,
		cmd:                 cmd,
	}

	return &res
}

// Run will run a powershell command and return the stdout and stderr
// The output is converted to JSON if the json parameter is set to true.
func (p *PSCommand) Run(conf *config.ProviderConf) (*PSCommandResult, error) {
	var (
		err      error
		exitCode int
		stderr   bytes.Buffer
		stdout   bytes.Buffer
	)
	conn, err := conf.AcquireSshClient()
	if err != nil {
		return nil, fmt.Errorf("while acquiring ssh client: %s", err)
	}
	defer conf.ReleaseSshClient(conn)

	encodedCmd := winrm.Powershell(p.cmd)

	cmd, err := conn.Command(encodedCmd)
	if err != nil {
		return nil, err
	}

	cmd.Session.Stderr = &stderr
	cmd.Session.Stdout = &stdout

	err = cmd.Run()
	if err != nil {
		if v, ok := err.(*ssh.ExitError); ok {
			exitCode = v.ExitStatus()
		} else {
			return nil, fmt.Errorf("run error: %s", err)
		}
	}

	out := stdout.String()
	if p.ForceArray && stdout.String() != "" && stdout.String()[0] != '[' {
		out = fmt.Sprintf("[%s]", stdout.String())
	}

	result := &PSCommandResult{
		Stdout:   out,
		StdErr:   stderr.String(),
		ExitCode: exitCode,
	}
	return result, nil
}

func (p *PSCommand) String() string {
	return p.cmd
}

// PSCommandResult holds the stdout, stderr and exit code of a powershell command
type PSCommandResult struct {
	Stdout   string
	StdErr   string
	ExitCode int
}
