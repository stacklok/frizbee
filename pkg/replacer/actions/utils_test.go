package actions

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stacklok/frizbee/pkg/ghrest"
)

func TestGetChecksum(t *testing.T) {
	t.Parallel()

	tok := os.Getenv("GITHUB_TOKEN")

	type args struct {
		action string
		ref    string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "actions/checkout with v4.1.1",
			args: args{
				action: "actions/checkout",
				ref:    "v4.1.1",
			},
			want:    "b4ffde65f46336ab88eb53be808477a3936bae11",
			wantErr: false,
		},
		{
			name: "actions/checkout with v3.6.0",
			args: args{
				action: "actions/checkout",
				ref:    "v3.6.0",
			},
			want:    "f43a0e5ff2bd294095638e18286ca9a3d1956744",
			wantErr: false,
		},
		{
			name: "actions/checkout with checksum returns checksum",
			args: args{
				action: "actions/checkout",
				ref:    "1d96c772d19495a3b5c517cd2bc0cb401ea0529f",
			},
			want:    "1d96c772d19495a3b5c517cd2bc0cb401ea0529f",
			wantErr: false,
		},
		{
			name: "aquasecurity/trivy-action with 0.14.0",
			args: args{
				action: "aquasecurity/trivy-action",
				ref:    "0.14.0",
			},
			want:    "2b6a709cf9c4025c5438138008beaddbb02086f0",
			wantErr: false,
		},
		{
			name: "aquasecurity/trivy-action with branch returns checksum",
			args: args{
				action: "aquasecurity/trivy-action",
				ref:    "bump-trivy",
			},
			want:    "fb5e1b36be448e92ca98648c661bd7e9da1f1317",
			wantErr: false,
		},
		{
			name: "actions/checkout with invalid tag returns error",
			args: args{
				action: "actions/checkout",
				ref:    "v4.1.1.1",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "actions/checkout with invalid action returns error",
			args: args{
				action: "invalid-action",
				ref:    "v4.1.1",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "actions/checkout with empty action returns error",
			args: args{
				action: "",
				ref:    "v4.1.1",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "actions/checkout with empty tag returns error",
			args: args{
				action: "actions/checkout",
				ref:    "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "bufbuild/buf-setup-action with v1 is an array",
			args: args{
				action: "bufbuild/buf-setup-action",
				ref:    "v1",
			},
			want: "f0475db2e1b1b2e8d121066b59dfb7f7bd6c4dc4",
		},
		{
			name: "anchore/sbom-action/download-syft with a sub-action works",
			args: args{
				action: "anchore/sbom-action/download-syft",
				ref:    "v0.14.3",
			},
			want: "78fc58e266e87a38d4194b2137a3d4e9bcaf7ca1",
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ghcli := ghrest.NewClient(tok)
			got, err := GetChecksum(context.Background(), ghcli, tt.args.action, tt.args.ref)
			if tt.wantErr {
				require.Error(t, err, "Wanted error, got none")
				require.Empty(t, got, "Wanted empty string, got %v", got)
				return
			}
			require.NoError(t, err, "Wanted no error, got %v", err)
			require.Equal(t, tt.want, got, "Wanted %v, got %v", tt.want, got)
		})
	}
}
