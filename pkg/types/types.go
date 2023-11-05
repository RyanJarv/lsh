package types

import (
	"fmt"
	"github.com/aws/constructs-go/constructs/v10"
	"io"
	"reflect"
)

func NewRegistry(steps ...interface{}) Registry {
	r := Registry{}
	for _, step := range steps {
		f := reflect.TypeOf(step)
		if f.Kind() != reflect.Func {
			panic("step must be func")
		}

		out := reflect.New(f.Out(0).Elem())
		value := out.Interface()

		if v, ok := value.(IStep); !ok {
			panic(fmt.Sprintf("step must return CdkStep, got: %T %+v", value, value))
		} else {
			r[v.GetName()] = v
		}
	}

	return r
}

type Registry map[string]IStep

type IStep interface {
	GetName() string
}

type Step struct {
	Name  string
	Value interface{}
}

func (s Step) GetName() string {
	return s.Name
}

type CdkStep interface {
	Compile(constructs.Construct, interface{}, int) ([]interface{}, error)
}

type CdkStepFanOut interface {
	Compile(constructs.Construct, []interface{}, int) ([]interface{}, error)
}

type Process interface {
	io.ReaderFrom
	io.WriterTo
	Run() error
}

type Deployable interface {
	Deploy() error
}
