package messaging

import "errors"

type codecStub struct {
	marshalErr   error
	unmarshalErr error
}

func (c codecStub) Marshal(value any) ([]byte, error) {
	if c.marshalErr != nil {
		return nil, c.marshalErr
	}

	return []byte("encoded"), nil
}

func (c codecStub) Unmarshal(_ []byte, _ any) error {
	if c.unmarshalErr != nil {
		return c.unmarshalErr
	}

	return nil
}

func (c codecStub) ContentType() string {
	return ContentTypeJSON
}

type stubCodec struct{}

func (stubCodec) Marshal(any) ([]byte, error) {
	return []byte("encoded"), nil
}

func (stubCodec) Unmarshal([]byte, any) error {
	return nil
}

func (stubCodec) ContentType() string {
	return ContentTypeJSON
}

func wrapDecodeError(err error) error {
	return errors.Join(ErrDecodeFailed, err)
}
