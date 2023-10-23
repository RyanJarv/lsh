package state

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	sfn "github.com/aws/aws-cdk-go/awscdk/v2/awsstepfunctions"
	"github.com/aws/aws-cdk-go/awscdk/v2/cxapi"
	"github.com/aws/jsii-runtime-go"
	L "github.com/ryanjarv/msh/pkg/logger"
	"github.com/ryanjarv/msh/pkg/types"
	"github.com/ryanjarv/msh/pkg/utils"
	"io"
	"os"
	"os/exec"
)

func (s *App) Run(step types.IStep) error {
	s.State.AddStep(step)

	if utils.IsTTY(os.Stdout) || os.Getenv("MSH_BUILD") != "" {
		return s.Build()
	} else {
		return s.State.WriteState()
	}
}

func (s *State) WriteState() error {
	d, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("writeState: failed to marshal state: %w", err)
	}

	_, err = io.WriteString(os.Stdout, string(d)+"\n")
	if err != nil {
		return fmt.Errorf("writeState: failed to write state: %w", err)
	}

	return nil
}

func (s *App) Build() error {
	app := awscdk.NewApp(&awscdk.AppProps{})
	stack := awscdk.NewStack(app, jsii.String("msh"), &awscdk.StackProps{})

	chain := sfn.Chain_Start(
		sfn.NewPass(stack, jsii.String("choice"), &sfn.PassProps{}),
	)

	for _, step := range s.State.Steps {
		if s, ok := step.Value.(types.CdkStep); ok {
			s.CdkStep(stack)
		}

		if s, ok := step.Value.(types.SfnStep); ok {
			chain = s.SfnHook(stack, chain)
		}
	}

	sfn.NewStateMachine(stack, jsii.String("StateMachine"), &sfn.StateMachineProps{
		DefinitionBody: sfn.DefinitionBody_FromChainable(chain),
		Timeout:        awscdk.Duration_Minutes(jsii.Number(5)),
		Comment:        jsii.String("a super cool state machine"),
	})

	synth := app.Synth(nil)
	if synth == nil || synth.Stacks() == nil || len(*synth.Stacks()) != 1 {
		return fmt.Errorf("build: failed to synthesize app: %s", synth)
	}

	L.Debug.Println("synth directory:", *synth.Directory())

	err := Deploy(synth)
	if err != nil {
		return fmt.Errorf("build: failed to deploy: %w", err)
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