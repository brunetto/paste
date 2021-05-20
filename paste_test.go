package paste_test

import (
	"bytes"
	"io"
	"regexp"
	"testing"

	"github.com/brunetto/paste"
	"github.com/stretchr/testify/assert"
)

func TestReplace(t *testing.T) {
	t.Parallel()

	rpl := paste.FakeReplacer

	type args struct {
		rpl paste.Replacer
		in  io.Reader
	}
	tests := []struct {
		name string
		args args
		want string
		fail bool
	}{
		{
			name: "ok single line",
			args: args{rpl: rpl, in: bytes.NewBuffer([]byte("before <%/replace me%> after"))},
			want: "before --- after\n",
		},
		{
			name: "ok multi line",
			args: args{rpl: rpl, in: bytes.NewBuffer([]byte("line 1: before <%/replace me%> after\nline 2: before <%/replace me%> after"))},
			want: "line 1: before --- after\nline 2: before --- after\n",
		},
	}

	for _, tt := range tests {
		tt := tt // avoid messing with loop variable

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			out := &bytes.Buffer{}
			err := paste.ReplaceAll(tt.args.rpl, tt.args.in, out)
			if tt.fail {
				assert.NotNil(t, err)
				return
			}

			assert.Equal(t, tt.want, out.String())
		})
	}
}

func TestGetPlaceholder(t *testing.T) {
	t.Parallel()

	type args struct {
		rgx  *regexp.Regexp
		line string
	}
	tests := []struct {
		name   string
		args   args
		want   string
		exists bool
	}{
		{
			name:   "ok",
			args:   args{rgx: regexp.MustCompile(paste.RgxStr), line: "before <%/replace me%> after"},
			want:   "/replace me",
			exists: true,
		},
		{
			name:   "ok empty",
			args:   args{rgx: regexp.MustCompile(paste.RgxStr), line: "before <%%> after"},
			want:   "",
			exists: true,
		},
		{
			name:   "no placeholder",
			args:   args{rgx: regexp.MustCompile(paste.RgxStr), line: "before after"},
			want:   "",
			exists: false,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, exists := paste.GetPlaceholder(tt.args.rgx, tt.args.line)
			assert.Equal(t, tt.exists, exists)
			assert.Equal(t, tt.want, got)
		})
	}
}
