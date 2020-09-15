// Copyright (C) 2016-2018 Betalo AB
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

const version = "1.3.0"

func main() {
	var (
		forceFlag    = flag.Bool("f", false, "Force running the command even after giving up")
		infileFlag   = flag.String("i", "", "Read resources from file, '-' to read from stdin")
		quietFlag    = flag.Bool("q", false, "Set quiet mode")
		timeoutFlag  = flag.Duration("t", 1*time.Minute, "Set timeout duration before giving up")
		verbose1Flag = flag.Bool("v", false, "Set verbose output mode")
		verbose2Flag = flag.Bool("vv", false, "Set more verbose output mode")
		versionFlag  = flag.Bool("V", false, "Show version")
	)
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: await [options...] <res>... [ -- <cmd>]")
		fmt.Fprintln(os.Stderr, "Await availability of resources.")
		fmt.Fprintln(os.Stderr)
		flag.PrintDefaults()
	}
	flag.Parse()

	if *versionFlag {
		fmt.Println("await", version)
		return
	}

	var logLevel int
	switch {
	case *verbose2Flag:
		logLevel = debugLevel
	case *verbose1Flag:
		logLevel = infoLevel
	case *quietFlag:
		logLevel = silentLevel
	default:
		logLevel = errorLevel
	}
	logger := NewLogger(logLevel)

	resArgs, cmdArgs := splitArgs(flag.Args())
	if *infileFlag != "" {
		resFile, err := readFromFile(*infileFlag)
		if err != nil {
			logger.Fatalf("Error: failed to read resources file: %v", err)
		}
		resArgs = append(resArgs, resFile...)
	}
	ress, err := parseResources(resArgs)
	if err != nil {
		logger.Fatalf("Error: failed to parse resources: %v", err)
	}

	awaiter := &awaiter{
		logger:  logger,
		timeout: *timeoutFlag,
	}

	if err := awaiter.run(ress); err != nil {
		if e, ok := err.(*unavailabilityError); ok {
			logger.Errorf("Resource unavailable: %v", e)
			logger.Errorln("Timeout exceeded")
		} else {
			logger.Fatalf("Error: %v", err)
		}
		if !*forceFlag {
			os.Exit(1)
		}
	} else {
		logger.Infoln("All resources available")
	}

	if len(cmdArgs) > 0 {
		logger.Infof("Running command: %v", cmdArgs)
		if err := execCmd(cmdArgs); err != nil {
			logger.Fatalf("Error: failed to execute command: %v", err)
		}
	}
}

func splitArgs(args []string) ([]string, []string) {
	if i := indexOf(args, "--"); i >= 0 {
		return args[0:i], args[i+1:]
	}

	// We haven't seen a resource|command separator ('--'). This can either be
	// because of Go's flag parser removing the separator if no args were given,
	// or because there actually was none given.
	// Fallback to the original (unparsed) flag argument list and see if a
	// separator was given there and if, assume all arguments given are part of
	// the command, i.e. no resources at all were provided.
	if i := indexOf(os.Args, "--"); i >= 0 {
		return []string{}, args
	}

	// We still haven't seen a resource|command separator ('--'). Now finally
	// assume none was given and use all arguments as resources.
	return args, []string{}
}

func indexOf(l []string, s string) int {
	for i, e := range l {
		if e == s {
			return i
		}
	}
	return -1
}

func stdinFromPipe() bool {
	stat, _ := os.Stdin.Stat()
	return (stat.Mode() & os.ModeNamedPipe) != 0
}

func readFromFile(filepath string) ([]string, error) {
	f := os.Stdin
	if filepath != "-" {
		var err error
		if f, err = os.Open(filepath); err != nil {
			return nil, err
		}
		defer func() { _ = f.Close() }()
	}

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}

func execCmd(cmdArgs []string) error {
	path, err := exec.LookPath(cmdArgs[0])
	if err != nil {
		return err
	}
	return syscall.Exec(path, cmdArgs, os.Environ())
}
