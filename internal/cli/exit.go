package cli

import (
	"errors"

	"github.com/eremenko789/singctl/internal/api"
)

// ExitCode maps an error to a process exit code for scriptability.
// nil → 0; NotFound → 3; Config → 2; all other errors (including CLI misuse) → 1.
// Public table: docs/scriptability.md and specs/007-scriptability-exits/contracts/exit-codes-public.md.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	var ce *api.ClassifiedError
	if !errors.As(err, &ce) {
		classified := api.Classify(err)
		if !errors.As(classified, &ce) {
			return 1
		}
	}

	switch ce.Kind {
	case api.KindNotFound:
		return 3
	case api.KindConfig:
		return 2
	default:
		return 1
	}
}
