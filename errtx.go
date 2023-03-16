// ErrTX is a library that offers error tracing capabilities,
// all while prioritizing performance and minimalism.
// ErrTX is a library that allows for error tracking while striving to maintain a high level of efficiency and minimalism.
package errtx

import (
	"errors"
	"fmt"
	"runtime"
)

type tracedError struct {
	err   error
	spans []errorSpan
}

type errorSpan struct {
	msg string
	loc string
}

func getLocation() string {
	_, file, line, _ := runtime.Caller(2)
	return fmt.Sprintf("%s:%d", file, line)
}

// Is returns whether the error is another error for use with errors.Is.
func (e tracedError) Is(target error) bool {
	return errors.Is(e.err, target)
}

// As returns the error as another error for use with errors.As.
func (e tracedError) As(target interface{}) bool {
	return errors.As(e.err, target)
}

// Unwrap returns the wrapped error for use with errors.Unwrap.
func (e tracedError) Unwrap() error {
	return errors.Unwrap(e.err)
}

// Error implements the error interface.
func (e tracedError) Error() string {
	return e.err.Error()
}

// Format implements fmt.Formatter.
func (e tracedError) Format(f fmt.State, verb rune) {
	if verb == 'v' && f.Flag('+') {
		fmt.Fprint(f, e.traceRepr())
		return
	}
	fmt.Fprintf(f, fmt.FormatString(f, verb), e.err)
}

func (e tracedError) traceRepr() string {
	if len(e.spans) == 0 {
		return ""
	}

	lenRepr := 0
	for _, span := range e.spans {
		if span.msg != "" {
			lenRepr += len(span.msg) + 1
		}
		if span.loc != "" {
			lenRepr += len(span.loc) + 2
		}
	}

	buffer := make([]byte, 0, lenRepr)
	for i := len(e.spans) - 1; i >= 0; i-- {
		if e.spans[i].msg != "" {
			buffer = append(buffer, []byte(e.spans[i].msg)...)
			buffer = append(buffer, '\n')
		}
		if e.spans[i].loc != "" {
			buffer = append(buffer, '\t')
			buffer = append(buffer, []byte(e.spans[i].loc)...)
			buffer = append(buffer, '\n')
		}
	}

	return string(buffer[:len(buffer)-1])
}

func Newf(format string, a ...any) error {
	err := fmt.Errorf(format, a...)
	spans := []errorSpan{
		{
			msg: err.Error(),
			loc: getLocation(),
		},
	}

	return tracedError{err, spans}
}

func Wrapf(err error, format string, a ...any) error {
	if err == nil {
		return nil
	}
	newErr := fmt.Errorf(format, a...)

	te, ok := err.(tracedError)
	if ok {
		te.spans = append(te.spans, errorSpan{
			msg: newErr.Error(),
			loc: getLocation(),
		})
		te.err = fmt.Errorf("%w - %w", newErr, te.err)
		return te
	}

	spans := []errorSpan{
		{
			msg: err.Error(),
			loc: "",
		},
		{
			msg: newErr.Error(),
			loc: getLocation(),
		},
	}

	return tracedError{fmt.Errorf("%w - %w", newErr, err), spans}
}

func Trace(err error) error {
	te, ok := err.(tracedError)
	if ok {
		te.spans = append(te.spans, errorSpan{
			msg: "",
			loc: getLocation(),
		})
		return te
	}

	spans := []errorSpan{
		{
			msg: err.Error(),
			loc: "",
		},
		{
			msg: "",
			loc: getLocation(),
		},
	}

	return tracedError{err, spans}
}
