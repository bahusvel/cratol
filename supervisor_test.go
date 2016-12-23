package cratol

import (
	"log"
	"testing"
)

var counter = 0

func Crashable(message string) {
	if counter < 10 {
		counter++
		panic(message)
	}
}

func TestRestart(t *testing.T) {
	pid, err := GlobalSupervisor.RunTask(Crashable, "Hello")
	if err != nil {
		t.Error(err)
	}
	log.Println("Pid", pid)
	err = GlobalSupervisor.AwaitTask(pid)
	if err != nil {
		t.Error(err)
	}
}
