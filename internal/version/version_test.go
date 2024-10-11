package version

import (
	"testing"
)

func TestVersion_Next(t *testing.T) {
	type fields struct {
		version string
	}
	type args struct {
		t   ChangeType
		pre string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "not a number major",
			fields: fields{
				version: "va.1.0",
			},
			args: args{
				t: ChangeTypeMajor,
			},
			want:    "0.0.0",
			wantErr: true,
		},
		{
			name: "missing build version",
			fields: fields{
				version: "v0.1.0-alpha",
			},
			args: args{
				t: ChangeTypeMajor,
			},
			want:    "0.0.0",
			wantErr: true,
		},
		{
			name: "bumps major without build",
			fields: fields{
				version: "v1.2.3",
			},
			args: args{
				t: ChangeTypeMajor,
			},
			want:    "v2.0.0",
			wantErr: false,
		},
		{
			name: "bumps minor without build",
			fields: fields{
				version: "v1.2.3",
			},
			args: args{
				t: ChangeTypeMinor,
			},
			want:    "v1.3.0",
			wantErr: false,
		},
		{
			name: "bumps patch without build",
			fields: fields{
				version: "v1.2.3",
			},
			args: args{
				t: ChangeTypePatch,
			},
			want:    "v1.2.4",
			wantErr: false,
		},
		{
			name: "adds pre-release suffix and bumps version",
			fields: fields{
				version: "v1.2.3",
			},
			args: args{
				t:   ChangeTypePatch,
				pre: "alpha",
			},
			want:    "v1.2.4-alpha.1",
			wantErr: false,
		},
		{
			name: "bumps build version",
			fields: fields{
				version: "v1.2.3-alpha.1",
			},
			args: args{
				t:   ChangeTypePatch,
				pre: "alpha",
			},
			want:    "v1.2.3-alpha.2",
			wantErr: false,
		},
		{
			name: "bumps build version",
			fields: fields{
				version: "v1.2.3-alpha.1",
			},
			args: args{
				t:   ChangeTypeNone,
				pre: "alpha",
			},
			want:    "v1.2.3-alpha.2",
			wantErr: false,
		},
		{
			name: "changes pre-release",
			fields: fields{
				version: "v1.2.3-alpha.2",
			},
			args: args{
				t:   ChangeTypePatch,
				pre: "beta",
			},
			want:    "v1.2.3-beta.1",
			wantErr: false,
		},
		{
			name: "changes minor with the same pre-release",
			fields: fields{
				version: "v1.2.3-alpha.2",
			},
			args: args{
				t:   ChangeTypeMinor,
				pre: "alpha",
			},
			want:    "v1.3.0-alpha.1",
			wantErr: false,
		},
		{
			name: "changes minor with different pre-release",
			fields: fields{
				version: "v1.2.3-alpha.2",
			},
			args: args{
				t:   ChangeTypeMinor,
				pre: "beta",
			},
			want:    "v1.3.0-beta.1",
			wantErr: false,
		},
		{
			name: "changes major without pre-release",
			fields: fields{
				version: "v1.2.3-alpha.2",
			},
			args: args{
				t: ChangeTypeMajor,
			},
			want:    "v2.0.0",
			wantErr: false,
		},
		{
			name: "changes minor with pre-release",
			fields: fields{
				version: "v2.0.0",
			},
			args: args{
				t:   ChangeTypeMinor,
				pre: "pre",
			},
			want:    "v2.1.0-pre.1",
			wantErr: false,
		},
		{
			name: "drops pre-release on patch version bump",
			fields: fields{
				version: "v1.2.3-alpha.2",
			},
			args: args{
				t:   ChangeTypePatch,
				pre: "",
			},
			want:    "v1.2.3",
			wantErr: false,
		},
		{
			name: "drops pre-release on minor version bump",
			fields: fields{
				version: "v1.2.3-alpha.2",
			},
			args: args{
				t:   ChangeTypeMinor,
				pre: "",
			},
			want:    "v1.3.0",
			wantErr: false,
		},
		{
			name: "drops pre-release on major version bump",
			fields: fields{
				version: "v1.2.3-alpha.2",
			},
			args: args{
				t:   ChangeTypeMajor,
				pre: "",
			},
			want:    "v2.0.0",
			wantErr: false,
		},
		{
			name: "drops pre-release on no change type",
			fields: fields{
				version: "v1.2.3-alpha.2",
			},
			args: args{
				t:   ChangeTypeNone,
				pre: "",
			},
			want:    "v1.2.3",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := New(tt.fields.version)
			if (err != nil) != tt.wantErr {
				t.Fatalf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got := v.Next(tt.args.t, tt.args.pre).String(); got != tt.want {
				t.Fatalf("Version.Next() = %v, want %v", got, tt.want)
			}
		})
	}
}
