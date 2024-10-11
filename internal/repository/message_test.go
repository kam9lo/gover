package repository

import (
	"reflect"
	"testing"
)

func TestParseMessage(t *testing.T) {
	type args struct {
		template string
		message  string
		required []string
	}
	tests := []struct {
		name    string
		args    args
		want    Message
		wantErr bool
	}{
		{
			name: "parses exact format",
			args: args{
				template: `{{.Type}}({{.Scope}}): {{.Message}}

{{.Description}}`,
				message: `improvement(semver): commit message

description`,
				required: []string{"Type", "Scope", "Message", "Description"},
			},
			want: Message{
				"Type":        "improvement",
				"Scope":       "semver",
				"Message":     "commit message",
				"Description": "description",
			},
		},
		{
			name: "parses on missing description field",
			args: args{
				template: "{{.Type}}({{.Scope}}): {{.Message}}\n\n{{.Description}}",
				message:  `improvement(semver): commit message`,
				required: []string{"Type", "Scope", "Message"},
			},
			want: Message{
				"Type":    "improvement",
				"Scope":   "semver",
				"Message": "commit message",
			},
		},
		{
			name: "fails on missing required field",
			args: args{
				template: "{{.Type}}({{.Scope}}): {{.Message}}\n\n{{.Description}}",
				message:  `improvement(semver): commit message`,
				required: []string{"Type", "Scope", "Message", "Description"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "parses on missing template match",
			args: args{
				template: ``,
				message:  `description`,
				required: []string{},
			},
			want: Message{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMessage(
				tt.args.template, tt.args.message, tt.args.required...,
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMessage() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseMessage() = %v, want %v (err: %v)", got, tt.want, err)
			}
		})
	}
}
