package model

type ShutdownReport struct {
	// shutdown resource names
	DoneResources []string
	// already stopped resource names
	AlreadyShutdownResources []string
}
