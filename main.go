// gccgo -Wall -static -g -pg main.go
// comunicacion hijo -> padre

package main

import "fmt"
import "os"
import "syscall"


//extern syscall.runtime_BeforeFork
func syscall_runtime_BeforeFork()

//extern syscall.runtime_AfterFork
func syscall_runtime_AfterFork()

//extern syscall.runtime_AfterForkInChild
func syscall_runtime_AfterForkInChild()

var _step int

func step() {
	_step++
	fmt.Printf("step  %3d %d\n", _step, os.Getpid())
}

type info struct {
	detail   string
	code     uint8
}

func main() {

	fmt.Println("init     ", os.Getpid())
	step()

        var fd [2]int

	fd[0] = -1
	fd[1] = -1

	err := syscall.Pipe(fd[:])
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("pipe fd  ", fd)

	syscall.ForkLock.Lock()

	syscall_runtime_BeforeFork()

	defer func() {
		fmt.Println("exit     ", os.Getpid())
	}()

	pid, _, err1 := syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)

	if err1 != 0 {
		syscall_runtime_AfterFork()
		syscall.ForkLock.Unlock()
		fmt.Println(syscall.Errno(err1))
		return
	}

	if pid != 0 {
		step()
		syscall_runtime_AfterFork()
		syscall.ForkLock.Unlock()

		syscall.Close(fd[1]) // Cierra el descriptor de escritura

		buf := make([]byte, 1)
		syscall.Read(fd[0], buf)

		fmt.Println("close", syscall.Close(fd[0]))

		fmt.Println(buf)

		step()
		return
	}

	syscall_runtime_AfterForkInChild()

	syscall.Close(fd[0]) // Cierra el descriptor de lectura
	syscall.Write(fd[1], []byte{129})
	fmt.Println("close", syscall.Close(fd[1]))
	step()

	for i := 0; i < 20; i++ {
		go func() {
			fmt.Println("goroutine, no work in child")
			step()
		}()
	}

	for i := 0; i < 15; i++ {
		step()
	}

	// sleep no work in child
	// time.Sleep(10 * time.Second)

	return
}
