package hub

import (
	"errors"
	"testing"

	"github.com/vortexui/vortexui/internal/domain"
)

func TestClassifyDialError(t *testing.T) {
	cases := []struct {
		err  error
		want domain.NodeDiagCode
	}{
		{errors.New("authentication handshake failed: x509: certificate signed by unknown authority"), domain.NodeDiagMTLS},
		{errors.New("dial tcp 1.2.3.4:50051: connect: connection refused"), domain.NodeDiagUnreachable},
		{errors.New("context deadline exceeded"), domain.NodeDiagUnreachable},
		{nil, domain.NodeDiagOK},
	}
	for _, tc := range cases {
		got := classifyDialError(tc.err)
		if got.Code != tc.want {
			t.Fatalf("classify(%v) = %q, want %q", tc.err, got.Code, tc.want)
		}
	}
}
