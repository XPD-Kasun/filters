package filters

import "errors"

// Denotes that the given operation is not supported by this visitor.
var ErrUnsupported error = errors.New("Unsupported")
