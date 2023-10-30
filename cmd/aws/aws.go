package aws

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsstepfunctions"
	tasks "github.com/aws/aws-cdk-go/awscdk/v2/awsstepfunctionstasks"
	"github.com/aws/aws-cdk-go/awscdk/v2/lambdalayerawscli"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/jsii-runtime-go"
	"github.com/samber/lo"
	"os"
	"os/exec"
)

//go:embed cli.py
var code []byte

func New(args []string) (*LambdaCmd, error) {
	if lo.Contains(args, "--help") || lo.Contains(args, "help") {
		Help(args)
		os.Exit(1)
	}

	script, err := os.ReadFile(args[0])
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	if !bytes.Contains(script, []byte("lambda_handler")) {
		return nil, fmt.Errorf("script must contain a `lambda_handler` function")
	}

	iamActions, err := IamActionsFromCliArgs(args)
	if err != nil {
		return nil, fmt.Errorf("getting iam actions from cli args: %w", err)
	}

	iamStatement := awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions:   jsii.Strings(iamActions...),
		Resources: jsii.Strings("*"),
	})
	return &LambdaCmd{
		IamStatement: []awsiam.PolicyStatement{iamStatement},
		Script:       string(code),
		Args:         args,
	}, nil
}

func Help(args []string) {
	cmd := exec.Command("aws", args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Run()
}

type LambdaCmd struct {
	Script       string
	Args         []string
	IamStatement []awsiam.PolicyStatement
}

func (s LambdaCmd) GetName() string { return "lambda" }

func (s LambdaCmd) Compile(stack awscdk.Stack, next interface{}) ([]interface{}, error) {
	function := awslambda.NewFunction(stack, jsii.String("awscli"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_PYTHON_3_11(),
		Handler: jsii.String("index.lambda_handler"),
		Code:    awslambda.Code_FromInline(jsii.String(s.Script)),
	})
	function.AddLayers(lambdalayerawscli.NewAwsCliLayer(stack, jsii.String("AwsCliLayer")))

	function.Role().AttachInlinePolicy(awsiam.NewPolicy(stack, jsii.String("AwsCliPolicy"), &awsiam.PolicyProps{
		Statements: &s.IamStatement,
	}))

	var this awsstepfunctions.INextable

	payload, err := json.Marshal(map[string]interface{}{"command": s.Args[1:]})
	if err != nil {
		return nil, fmt.Errorf("marshalling payload: %w", err)
	}

	this = tasks.NewLambdaInvoke(stack, jsii.String(fmt.Sprintf("invoke %s %d", flag.Arg(0), os.Getpid())), &tasks.LambdaInvokeProps{
		LambdaFunction: function,
		Payload:        awsstepfunctions.TaskInput_FromText(aws.String(string(payload))),
		OutputPath:     jsii.String("$.Payload"),
	})

	if next != nil {
		chain, ok := next.(awsstepfunctions.IChainable)
		if !ok {
			return nil, fmt.Errorf("next step must be statemachine chain")
		}

		this = this.Next(chain)
	}

	return []interface{}{this}, nil
}