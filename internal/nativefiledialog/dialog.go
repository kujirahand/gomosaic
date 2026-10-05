package nativefiledialog

import "errors"

// ErrUnsupported is returned only if this package is called on a platform
// without a native dialog implementation.
var ErrUnsupported = errors.New("native file dialogs are not supported on this platform")
