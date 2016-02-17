package k8s

// chanWriter is an io.Writer that immediately writes all received bytes to a chan<- string
type chanWriter struct {
	ch chan<- []byte
}

// Write is the io.Writer interface implementation
func (c chanWriter) Write(b []byte) (int, error) {
	c.ch <- b
	return len(b), nil
}
