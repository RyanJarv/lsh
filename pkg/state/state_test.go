package state

import (
	"github.com/ryanjarv/msh/pkg/providers/lambda"
	"reflect"
	"testing"
)

func TestReadState(t *testing.T) {
	type args struct {
		line []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *State
		wantErr bool
	}{
		{
			name: "should read state",
			args: args{
				line: []byte(`{"Steps":[{"Name":"lambda","Value":{"Script":"script","Args":["args"]}}]}`),
			},
			want: &State{
				Steps: []step{
					{
						Name:  "lambda",
						Value: &lambda.LambdaCmd{Script: "script", Args: []string{"args"}},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := App{}
			got, err := app.ReadState(tt.args.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadState() got = %v, want %v", got, tt.want)
			}
		})
	}
}