package main

import (
	"context"
	"flag"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/ryanjarv/msh/pkg/fd"
	L "github.com/ryanjarv/msh/pkg/logger"
	"github.com/ryanjarv/msh/pkg/providers/process"
	"github.com/samber/lo"
	"io"
)

func main() {
	flag.Parse()

	sqs := &Sqs{
		Name: aws.String("test-4919"),
	}

	proc := process.NewProcess(sqs)
	err := proc.Run()

	if err != nil {
		L.Error.Fatalln(err)
	}
}

type Sqs struct {
	fd.ISqs
	Name *string
}

func (s *Sqs) Run() error {
	s.ISqs.Wait()
	return nil
}

func (s *Sqs) SetStdin(p interface{}) {
	s.ISqs = lo.Must(fd.NewSqsFrom(context.TODO(), p, *s.Name, "stdin"))
}

func (s *Sqs) GetStdout() io.ReadCloser {
	return s.ISqs
}