package cratol

import (
	"fmt"
	"log"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"
)

const (
	STATE_RUNNING  = iota
	STATE_WAITING  = iota
	STATE_FINISHED = iota
)

var GlobalSupervisor = Supervisor{tasks: map[Pid]*Task{}}

var pidMutex = sync.Mutex{}
var globalID Pid = 1

type Pid int

type Task struct {
	TaskID       Pid
	routine      reflect.Value
	args         []reflect.Value
	State        int
	eventChannel chan struct{}
}

type Supervisor struct {
	tasks map[Pid]*Task
}

func (this *Supervisor) RunTask(routine interface{}, args ...interface{}) (Pid, error) {
	taskid := this.AddTask(routine, args...)
	err := this.StartTask(taskid)
	return taskid, err
}

func (this *Supervisor) AddTask(routine interface{}, args ...interface{}) Pid {
	task := Task{State: STATE_WAITING, eventChannel: make(chan struct{})}

	pidMutex.Lock()
	task.TaskID = globalID
	globalID++
	pidMutex.Unlock()

	val := reflect.ValueOf(routine)
	if val.Kind() != reflect.Func {
		panic(fmt.Errorf("%s is not a function", routine))
	}
	task.routine = val
	tmpArgs := []reflect.Value{}
	for _, arg := range args {
		tmpArgs = append(tmpArgs, reflect.ValueOf(arg))
	}
	task.args = tmpArgs
	this.tasks[task.TaskID] = &task
	return task.TaskID
}

func (this *Supervisor) StartTask(taskid Pid) error {
	task, ok := this.tasks[taskid]
	if !ok {
		return fmt.Errorf("Task %d does not exist", taskid)
	}
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovering task %d from %v\n", taskid, r)
				strace := debug.Stack()
				lines := strings.Split(string(strace), "\n")[5:]
				fmt.Print(strings.Join(lines, "\n"))
				this.StartTask(taskid)
			}
		}()
		task.State = STATE_RUNNING
		_ = task.routine.Call(task.args)
		task.State = STATE_FINISHED
		task.eventChannel <- struct{}{}
	}()
	return nil
}

func (this *Supervisor) AwaitTask(taskid Pid) error {
	task, ok := this.tasks[taskid]
	if !ok {
		return fmt.Errorf("Task %d does not exist", taskid)
	}
	if task.State == STATE_FINISHED {
		return nil
	}
	for _ = range task.eventChannel {
		if task.State == STATE_FINISHED {
			return nil
		}
	}
	return nil
}
