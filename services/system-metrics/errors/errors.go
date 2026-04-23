package errors

import systemerrors "errors"

// the total CPU counter did not advance between two samples.

//

// This typically means the samples are identical or too close in time to

// produce a meaningful delta.
var ErrInvalidCPUDelta = systemerrors.New("invalid cpu delta")
