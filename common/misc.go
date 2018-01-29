package common

import "io"

type PipePair struct {
	Reader *io.ReadCloser
	Writer *io.WriteCloser
}

func (this *PipePair) Read(p []byte) (int, error) {
	r := *this.Reader
	return r.Read(p)
}

func (this *PipePair) Write(p []byte) (int, error) {
	w := *this.Writer
	return w.Write(p)
}

func (this *PipePair) Close() error {
	r := *this.Reader
	w := *this.Writer
	w.Close()
	return r.Close()
}
