package errors

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//errorWriter is a writer of the errors produced by this
//package. By default errors are discarted
var errorWriter = ioutil.Discard

//SetLogger sets the default output for error logging.
//If this function is not called, errors are discarted
func SetLogger(w io.Writer) {
	errorWriter = w
}

func logError(e *Error) {
	_, file, line, ok := runtime.Caller(2) // logger + public function.
	if ok {
		// Truncate file name at last file name separator.
		if index := strings.LastIndex(file, "/"); index >= 0 {
			file = file[index+1:]
		} else if index = strings.LastIndex(file, "\\"); index >= 0 {
			file = file[index+1:]
		}
	} else {
		file = "???"
		line = 1
	}
	fmt.Fprintln(errorWriter, file, line, e.Error(), time.Now().Local())
}

//Error describes a failed action with an status code
type Error struct {
	code    int
	message string
}

//Error returns the message of the error plus
//its status code
func (e *Error) Error() string {
	return fmt.Sprintf("Status %d : %s", e.code, e.message)
}

//Code returns the status code of the error
func (e *Error) Code() int {
	return e.code
}

//WithStatus sets the given status code of this error
func (e *Error) WithStatus(status int) *Error {
	e.code = status
	return e
}

//New returns an error with the given message, and a zero status code
func New(message string) *Error {
	e := &Error{message: message}
	logError(e)
	return e
}

//WithStatus annotates a new error with an status code
func WithStatus(code int, err error) *Error {
	if err == nil {
		return nil
	}
	e := &Error{code: code, message: err.Error()}
	logError(e)
	return e
}

//WithMessage anotates the error with an optional
func WithMessage(message string, code int, err error) *Error {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf("%s: %s", message, err.Error())
	e := &Error{code: code, message: msg}
	logError(e)
	return e
}

//MarshalJSON satisfies the json.Marshaler interface
func (e *Error) MarshalJSON() ([]byte, error) {
	errMap := map[string]string{
		"error": e.message,
		"code":  strconv.Itoa(e.code),
	}
	return json.Marshal(errMap)
}

//StackError contains an error message with an status and the callees stack
type StackError struct {
	err Error
	*stack
}

//WithStack returns an error annotated with the stacktrace
func WithStack(message string, status int, err error) *StackError {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf("%s: %s", message, err.Error())
	ex := Error{code: status, message: msg}
	return &StackError{err: ex, stack: callers()}
}

//Error returns the annotated error with its status code and the stacktrace.
func (s *StackError) Error() string {
	return fmt.Sprintf("%v%+v", s.err.Error(), s.stack)
}

//Code returns the status code of the error
func (s *StackError) Code() int {
	return s.err.Code()
}

//WithStatus sets the given status to this error
func (s *StackError) WithStatus(status int) {
	s.err.code = status
}
