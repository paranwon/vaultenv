package masker

import "io"

// Writer wraps an io.Writer and masks any registered secrets before writing.
type Writer struct {
	masker *Masker
	underlying io.Writer
}

// NewWriter returns a Writer that masks secrets using m before forwarding
// bytes to w.
func NewWriter(w io.Writer, m *Masker) *Writer {
	return &Writer{
		masker:     m,
		underlying: w,
	}
}

// Write masks p and writes the result to the underlying writer.
// The number of bytes reported as written always equals len(p) so that
// callers (e.g. log.SetOutput) do not see spurious short-write errors when
// masked output is shorter than the original.
func (w *Writer) Write(p []byte) (int, error) {
	masked := w.masker.Mask(string(p))
	_, err := io.WriteString(w.underlying, masked)
	if err != nil {
		return 0, err
	}
	// Report the original length to satisfy io.Writer contract for callers
	// that compare n against len(p).
	return len(p), nil
}
