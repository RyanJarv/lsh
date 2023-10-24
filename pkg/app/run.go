package app

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/cxapi"
	"github.com/aws/jsii-runtime-go"
	L "github.com/ryanjarv/msh/pkg/logger"
	"github.com/ryanjarv/msh/pkg/types"
	"github.com/ryanjarv/msh/pkg/utils"
	"github.com/samber/lo"
	"io"
	"os"
	"os/exec"
)

func (a *App) Run(step types.IStep) error {
	a.State.AddStep(step)

	if utils.IsTTY(os.Stdout) || os.Getenv("MSH_BUILD") != "" {
		return a.Build()
	} else {
		return a.State.WriteState()
	}
}

func (s *State) WriteState() error {
	d, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("writeState: failed to marshal app: %w", err)
	}

	_, err = io.WriteString(os.Stdout, string(d)+"\n")
	if err != nil {
		return fmt.Errorf("writeState: failed to write app: %w", err)
	}

	return nil
}

func (a *App) Build() error {
	app := awscdk.NewApp(&awscdk.AppProps{})
	stack := awscdk.NewStack(app, jsii.String("msh"), &awscdk.StackProps{})

	// pipeline represents the output of the last step, which will be passed to the next.
	var next interface{}

	// Reverse the steps so the source receives the next step instead of the previous one.
	steps := lo.Reverse(a.State.Steps)

	for _, step := range steps {
		s, ok := step.Value.(types.CdkStep)
		if !ok {
			return fmt.Errorf("build: not a cdk step (check the registry?): %T: %+v", step.Value, step.Value)
		}

		var err error
		next, err = s.Run(stack, next)
		if err != nil {
			return fmt.Errorf("build: failed to run step: %w", err)
		}
	}

	synth := app.Synth(nil)
	if synth == nil || synth.Stacks() == nil || len(*synth.Stacks()) != 1 {
		return fmt.Errorf("build: failed to synthesize app: %v", synth)
	}

	for _, stack := range *synth.Stacks() {
		L.Debug.Println(string(lo.Must(json.Marshal(stack.Template()))))
	}

	L.Debug.Println("synth directory:", *synth.Directory())

	if os.Getenv("MSH_SKIPDEPLOY") == "" {
		err := Deploy(synth)
		if err != nil {
			return fmt.Errorf("build: failed to deploy: %w", err)
		}
	}

	return nil
}

func Deploy(synth cxapi.CloudAssembly) error {
	cmd := exec.Command("cdk", "bootstrap", "--app", *synth.Directory())
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build: failed to bootstrap: %w", err)
	}

	cmd = exec.Command("cdk", "deploy", "--require-approval=never", "--app", *synth.Directory())
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build: failed to bootstrap: %w", err)
	}

	return nil
}
