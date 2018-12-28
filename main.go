package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func main() {
	flag.Parse()
	flags := flag.Args()

	cmd := exec.Command("ghq", "list", "-p")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	repos := strings.Split(string(out), "\n")

	swg := &sync.WaitGroup{}
	swg.Add(len(repos))

	output := make(chan string)
	errors := make(chan string)

	for _, repo := range repos {
		go commandProcess(swg, repo, flags, output, errors)
	}

	count := len(repos)
	errorMessages := []string{}
	for {
		select {
		case result := <-output:
			{
				println(result)
				count--
			}
		case result := <-errors:
			{
				errorMessages = append(errorMessages, result)
				count--
			}
		}
		if count == 0 {
			break
		}
	}

	for _, v := range errorMessages {
		println(v)
	}
	swg.Wait()
}

func commandProcess(swg *sync.WaitGroup, repo string, cmdString []string, output chan<- string, errors chan<- string) {
	defer swg.Done()

	cmd := exec.Command("git", cmdString...)
	cmd.Dir = repo
	result, err := cmd.CombinedOutput()

	resultString := repo + "\n" + string(result)

	if err != nil {
		errors <- resultString
	} else {
		output <- resultString
	}
}
