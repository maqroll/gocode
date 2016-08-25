package main

import (
	//	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
)

const (
	DIRECTORY_WORKERS  = 4
	DIRECTORY_CLEANERS = 4
)

var pendingDirectories int32

func open(directory string) *os.File {
	dir, err := os.Open(directory)
	if err != nil {
		panic(err)
	}
	return dir
}

func removeTarget(c chan string, end chan int) {
	for dir := range c {
		log.Printf("Cleaning %s\n", dir)
		cmd := exec.Command("mvn", "clean")
		cmd.Dir = dir
		err := cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
		err = cmd.Wait()
		if err != nil {
			log.Printf("Cleaning at %s failed: %v\n", dir, err)
		}
	}
	end <- 1
}

/*
Esta función busca módulos (que no sub-módulos en los sub-directorios).
Cuando encuentra un pom.xml en un directorio detiene el procesamiento de esa rama.
*/
func processDirectory(c chan *os.File, end chan int, clean chan string) {

	for dir := range c {
		// Comprobar si hay un fichero pom.xml en el directorio
		// Si lo hay, procesarlo.
		// Si no lo hay, procesar todos sus subdirectorios.
		if fileList, err := dir.Readdir(-1); err == nil {
			var found bool = false
			for _, f := range fileList {
				if !f.IsDir() && (f.Name() == "pom.xml") {
					found = true
				}
			}

			if found {
				clean <- dir.Name()
			} else {
				for _, f := range fileList {
					if f.IsDir() {
						d := open(filepath.Join(dir.Name(), f.Name()))

						atomic.AddInt32(&pendingDirectories, 1)
						c <- d
					}
				}
			}
		} else {
			panic(err)
		}

		atomic.AddInt32(&pendingDirectories, -1)
		if atomic.CompareAndSwapInt32(&pendingDirectories, 0, 0) {
			close(c)
			close(clean)
		}
	}
	end <- 1
}

func main() {
	dir := open(".")

	c := make(chan *os.File, 1000)
	end := make(chan int)

	clean := make(chan string, 100)

	for i := 0; i < DIRECTORY_WORKERS; i++ {
		go processDirectory(c, end, clean)
	}

	for i := 0; i < DIRECTORY_CLEANERS; i++ {
		go removeTarget(clean, end)
	}

	atomic.AddInt32(&pendingDirectories, 1)
	c <- dir

	for i := 0; i < DIRECTORY_WORKERS+DIRECTORY_CLEANERS; i++ {
		<-end
	}
}
