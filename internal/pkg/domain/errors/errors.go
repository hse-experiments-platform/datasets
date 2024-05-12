package errors

type RevertableError struct {
	baseErr error
	reason  string
}

func (r *RevertableError) Error() string {
	return "revertable error: " + r.baseErr.Error()
}

func (r *RevertableError) GetReason() string {
	return r.reason
}

func (r *RevertableError) Unwrap() error {
	return r.baseErr
}

func NewRevertable(err error, reason string) error {
	return &RevertableError{
		baseErr: err,
		reason:  reason,
	}
}
