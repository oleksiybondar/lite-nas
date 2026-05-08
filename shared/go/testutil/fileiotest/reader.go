package fileiotest

// Reader is a minimal fileio.Reader test double that returns configured data
// or a configured error.
type Reader struct {
	Data []byte
	Err  error
}

func (r Reader) Read() ([]byte, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	return r.Data, nil
}
