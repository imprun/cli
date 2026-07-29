package impruncli

import (
	"io"

	"github.com/imprun/cli/internal/controlcli"
)

func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return controlcli.RunImprunWithCredentialStore(args, stdin, stdout, stderr, credentialStore{})
}
