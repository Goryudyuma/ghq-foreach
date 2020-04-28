package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type Result interface {
	getResult() string
}

type Success struct {
	result string
}

func (s Success) getResult() string {
	return s.result
}

type Fail struct {
	result string
	err    error
}

func (f Fail) getResult() string {
	return f.result
}
func (f Fail) getError() string {
	return f.err.Error()
}

func main() {
	parallelNumber := flag.Int("n", 1, "parallel number")
	flag.Parse()
	commands := flag.Args()

	cmd := exec.Command("ghq", "list", "-p")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	repos := strings.Split(string(out), "\n")

	input := make(chan string)
	output := make(chan Result)

	swg := &sync.WaitGroup{}
	swg.Add(*parallelNumber)

	for i := 0; i < *parallelNumber; i++ {
		go process(swg, commands, input, output)
	}

	swg.Add(1)
	go func() {
		defer swg.Done()
		for _, repo := range repos {
			input <- repo
		}
		close(input)
	}()

	go func() {
		swg.Wait()
		close(output)
	}()

	fails := make([]Fail, 0)

exitLoop:
	for {
		select {
		case result, ok := <-output:
			if ok {
				switch result.(type) {
				case Success:
					s := result.(Success)
					println(s.getResult())
				case Fail:
					fails = append(fails, result.(Fail))
				}
			} else {
				break exitLoop
			}
		}
	}

	println("--------------------")

	for _, f := range fails {
		println(f.getResult())
		println(f.getError())
	}
}

func process(swg *sync.WaitGroup, commands []string, input <-chan string, output chan<- Result) {
	defer swg.Done()

	for {
		select {
		case repo, ok := <-input:
			if ok {
				cmd := exec.Command("git", commands...)
				cmd.Dir = repo
				result, err := cmd.CombinedOutput()

				resultString := repo + "\n" + string(result)

				if err != nil {
					output <- Fail{resultString, err}
				} else {
					output <- Success{resultString}
				}
			} else {
				return
			}
		}
	}
}
