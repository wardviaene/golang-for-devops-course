package api

type RequestError struct {
	Body     string
	HTTPCode int
	Err      string
}

func (r RequestError) Error() string {
	return r.Err
}
