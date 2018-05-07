package balancer

type errNoServiceAvailable string

var _ error = errNoServiceAvailable("")

// Error returns error string
func (e errNoServiceAvailable) Error() string { return string(e) }

// newErrNoServiceAvailable returns new errNoServiceAvailable formatted with
// balancer name string
func newErrNoServiceAvailable(balancer string) error {
	return errNoServiceAvailable(balancer + ": no service available")
}

// IsErrNoServiceAvailable returns a boolean indicating whether the error reports
// that no service endpoints are available at this moment
func IsErrNoServiceAvailable(err error) bool {
	_, ok := err.(errNoServiceAvailable)
	return ok
}
