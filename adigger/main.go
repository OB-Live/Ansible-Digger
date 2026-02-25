package adigger

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"encoding/json"
)

// --- Logger ---
var (
	verbose  bool
	infoLog  = log.New(os.Stdout, "[INFO] ", 0)
	warnLog  = log.New(os.Stdout, "[WARN] ", 0)
	errorLog = log.New(os.Stderr, "[ERROR] ", 0)
)

func initLogger(isVerbose bool) {
	verbose = isVerbose
}

func infof(format string, v ...interface{}) {
	if verbose {
		infoLog.Printf(format, v...)
	}
}

func warnf(format string, v ...interface{}) {
	warnLog.Printf(format, v...)
}

func errorf(format string, v ...interface{}) {
	errorLog.Printf(format, v...)
}


func PrintJSON(obj interface{}) { 
	bytes, _ := json.MarshalIndent(obj, "\t", "\t") 
	infof(string(bytes)) 
}

// --- Main Execution ---
func Run() int {
	rolesPath := flag.String("roles-path", "./roles", "Path to Ansible roles directory")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	daemon := flag.Bool("daemon", false, "Run in daemon mode as a web server")
	flag.Parse()

	initLogger(*verbose)

	inputFiles := flag.Args()
	if len(inputFiles) == 0 {
		// Default behavior: find all playbook_*.yaml files in the current directory.
		matches, err := filepath.Glob("playbook_*.yaml")
		if err != nil {
			errorf("Failed to find default playbook files: %v", err)
			return 1
		}
		inputFiles = matches
	}

	if len(inputFiles) == 0 {
		warnf("No input files found. Usage: go run . <playbook1.yaml> <playbook2.yaml> ...")
		return 1
	}

	for _, inputFile := range inputFiles {
		infof("Starting Ansible Digger for playbook: %s", inputFile)

		yamlData, err := ioutil.ReadFile(inputFile)
		if err != nil {
			errorf("Failed to read input file %s: %v", inputFile, err)
			continue // Continue to the next file
		}

		p := NewParser()
		playbook, err := p.Parse(yamlData)
		if err != nil {
			errorf("Failed to parse playbook %s: %v", inputFile, err)
			continue
		}

		// Enrich playbook with data from roles
		roleScanner := NewRoleScanner(*rolesPath) // Use the roles-path flag
		if err := roleScanner.ScanAndEnrich(playbook); err != nil {
			errorf("Failed during role scanning for %s: %v", inputFile, err)
			continue
		}

		// // Log the parsed playbook structure as YAML for debugging
		// parsedYAML, err := yaml.Marshal(playbook)
		// if err != nil {
		// 	errorf("Failed to marshal parsed playbook to YAML: %v", err)
		// } else {
		// 	infof("--- Parsed Playbook Structure ---\n%s\n---------------------------------", string(parsedYAML))
		// }

		if len(playbook.Plays) == 0 || (len(playbook.Plays[0].Tasks) == 0 && len(playbook.Plays[0].PreTasks) == 0) {
			warnf("Parsing resulted in an empty playbook structure. The generated graph will be empty.")
			continue
		}

		dotOutput := Render(playbook)
		outputFile := inputFile + ".dot"
		if err := ioutil.WriteFile(outputFile, []byte(dotOutput), 0644); err != nil {
			errorf("Failed to write output file %s: %v", outputFile, err)
			continue
		}

		infof("Successfully generated DOT file: %s", outputFile)
		fmt.Printf("✅ Graph generated for %s at %s\n", inputFile, filepath.Clean(outputFile))
	}

	if *daemon {
		return runWebServer()
	}
	return 0
}

func runWebServer() int {
	port := "8080"
	// Serve all files from the current directory.
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	infof("Starting web server on http://localhost:%s", port)
	fmt.Printf("\n🚀 Server listening on http://localhost:%s\n", port)
	fmt.Println("   Serving index.html and .dot files from the current directory.")

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		errorf("Failed to start web server: %v", err)
		return 1
	}
	return 0
}


func CleanID (x string) string{
	 x = strings.ReplaceAll(x, ".", "_")
	 x = strings.ReplaceAll(x, "-", "_")
	 x = strings.ReplaceAll(x, " ", "_")
	 return x
}