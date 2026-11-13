package container

import (
	"testing"
	"time"
)

func TestParseManagedSandboxes(t *testing.T) {
	older := time.Date(2026, 11, 13, 9, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 11, 13, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name    string
		output  string
		want    []ManagedSandbox
		wantErr bool
	}{
		{name: "no containers", output: "", want: nil},
		{
			name: "newest first",
			output: "aaa111\tshellforge-old\texited\t2026-11-13 09:00:00 +0000 UTC\n" +
				"bbb222\tshellforge-new\trunning\t2026-11-13 10:30:00 +0000 UTC\n",
			want: []ManagedSandbox{
				{ID: "bbb222", Name: "shellforge-new", State: "running", CreatedAt: newer},
				{ID: "aaa111", Name: "shellforge-old", State: "exited", CreatedAt: older},
			},
		},
		{
			name:   "numeric zone name",
			output: "ccc333\tshellforge-tokyo\trunning\t2026-11-13 19:30:00 +0900 +09\n",
			want:   []ManagedSandbox{{ID: "ccc333", Name: "shellforge-tokyo", State: "running", CreatedAt: newer}},
		},
		{name: "missing field", output: "aaa111\tshellforge-old\trunning\n", wantErr: true},
		{name: "bad time", output: "aaa111\tshellforge-old\trunning\tyesterday\n", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseManagedSandboxes(test.output)
			if (err != nil) != test.wantErr {
				t.Fatalf("parseManagedSandboxes() error = %v, wantErr %v", err, test.wantErr)
			}
			if len(got) != len(test.want) {
				t.Fatalf("parseManagedSandboxes() = %d sandboxes, want %d", len(got), len(test.want))
			}
			for index := range got {
				if got[index].ID != test.want[index].ID || got[index].Name != test.want[index].Name ||
					got[index].State != test.want[index].State || !got[index].CreatedAt.Equal(test.want[index].CreatedAt) {
					t.Errorf("sandbox %d = %+v, want %+v", index, got[index], test.want[index])
				}
			}
		})
	}
}
