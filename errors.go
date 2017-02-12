package main

import "fmt"

//ShutdownMessageError is a way to tell the watcher channels to stop listening to changes and quit
type ShutdownMessageError struct{}

func (e *ShutdownMessageError) Error() string {
	return fmt.Sprintf("shutdown command received, time to stop")
}

//IsShutdownError returns true if the reported error is of type ShutdownMessageError
func IsShutdownError(err error) bool {
	if _, ok := err.(*ShutdownMessageError); ok {
		return true
	}
	return false
}
