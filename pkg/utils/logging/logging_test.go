package logging

import (
	"testing"

	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestSetupLogging(t *testing.T) {
	type args struct {
		cfg *config.Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "text log format",
			args: args{
				cfg: &config.Config{
					LogFormat: "text",
				},
			},
			wantErr: false,
		},
		{
			name: "json log format",
			args: args{
				cfg: &config.Config{
					LogFormat: "json",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid log format",
			args: args{
				cfg: &config.Config{
					LogFormat: "invalid",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := SetupLogging(tt.args.cfg)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
