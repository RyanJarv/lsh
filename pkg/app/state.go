package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/ryanjarv/msh/pkg/types"
	"io"
	"reflect"
)

func ReadState(f io.Reader, registry types.Registry) (State, error) {
	buf := bufio.NewReader(f)

	line, err := buf.ReadBytes('\n')
	if err != nil {
		return State{}, fmt.Errorf("readConf: failed to read line: %w", err)
	}

	return UnmarshalState(registry, line)
}

func UnmarshalState(registry types.Registry, line []byte) (State, error) {
	state := struct {
		Steps []struct {
			Name  string
			Value json.RawMessage
		}
	}{
		Steps: []struct {
			Name  string
			Value json.RawMessage
		}{},
	}

	err := json.Unmarshal(line, &state)
	if err != nil {
		return State{}, fmt.Errorf("readConf: failed to unmarshal: %w", err)
	}

	steps := []types.Step{}

	// Set the correct type for each Step.
	for _, step := range state.Steps {
		// Make sure we're copying a value and not a pointer.
		v := reflect.ValueOf(registry[step.Name])
		if v.Kind() == reflect.Ptr {
			v = reflect.Indirect(v)
		}
	
		value := reflect.New(v.Type()).Interface()

		err := json.Unmarshal(step.Value, &value)
		if err != nil {
			return State{}, fmt.Errorf("readConf: failed to unmarshal step: %w", err)
		}

		steps = append(steps, types.Step{
			Name:  step.Name,
			Value: value,
		})
	}

	return State{
		Steps: steps,
	}, nil
}

type State struct {
	Steps []types.Step
}

// AddStep is called by each individual Step that wants to add to the cumulative app.
func (s *State) AddStep(c types.IStep) {
	s.Steps = append(s.Steps, types.Step{
		Name:  c.GetName(),
		Value: c,
	})
}
