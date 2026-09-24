package digisign

import (
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		NewVerifier,
		func(v *Verifier) port.SignatureVerifier { return v },
		func(v *Verifier) port.PDFSignatureVerifier { return v },
		func(v *Verifier) port.Timestamper { return v },
	),
)
