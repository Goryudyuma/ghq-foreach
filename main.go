package main

import (
	"flag"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/labstack/gommon/log"
)

func main() {
	flag.Parse()
	flags := flag.Args()

	cmd := exec.Command("ghq", "list", "-p")
	out, err := cmd.Output()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	repos := strings.Split(string(out), "\n")

	swg := &sync.WaitGroup{}
	swg.Add(len(repos))

	output := make(chan string)

	for _, repo := range repos {
		go commandProcess(swg, repo, flags, output)
	}

	count := len(repos)
	for {
		select {
		case result := <-output:
			{
				println(result)
				count--
			}
		}
		if count == 0 {
			break
		}
	}
	swg.Wait()
}

func commandProcess(swg *sync.WaitGroup, repo string, cmdString []string, output chan<- string) {
	defer swg.Done()

	cmd := exec.Command("git", cmdString...)
	cmd.Dir = repo
	result, err := cmd.CombinedOutput()
	if err != nil {
		output <- repo + "\n" + err.Error()
	} else {
		output <- repo + "\n" + string(result)
	}
}
