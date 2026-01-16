package build

import "errors"

// ErrNilOpts used to be returned in case opts passed are nil.
// This can be expected in some cases.
var ErrNilOpts = errors.New("nil options provided")
