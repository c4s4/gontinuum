package main

/*
 * Build management stuff. This is used to build modules.
 */

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

// Build is the result of a build.
type Build struct {
	Module  ModuleConfig
	Success bool
	Output  string
}

// Builds is a list of builds of the configuration.
type Builds []Build

func (build Build) String() string {
	if build.Success {
		return fmt.Sprintf("%s: OK", build.Module.Name)
	} else {
		return fmt.Sprintf("%s: ERROR", build.Module.Name)
	}
}

// Success tells if a list of builds was a success (that is if all buils were
// successful).
func (builds Builds) Success() bool {
	for _, build := range builds {
		if !build.Success {
			return false
		}
	}
	return true
}

// String returns a string that represents success or failure.
func (builds Builds) String() string {
	if builds.Success() {
		return "SUCCESS"
	} else {
		return "FAILURE"
	}
}

// BuildModule is called to build a module, that is:
// - get the repository clone.
// - run command to build the module.
// If build command returns 0 (as of Unix standard), the build is a success, else
// this is a failure.
func BuildModule(module ModuleConfig, directory string) Build {
	fmt.Printf("Building '%s'... ", module.Name)
	moduleDir := path.Join(directory, module.Name)
	// go in build directory
	currentDir, err := os.Getwd()
	defer os.Chdir(currentDir)
	err = os.Chdir(directory)
	if err != nil {
		return Build{
			Module:  module,
			Success: false,
			Output:  err.Error(),
		}
	}
	// delete module directory if it already exists
	if _, err := os.Stat(moduleDir); err == nil {
		os.RemoveAll(moduleDir)
	}
	// get the module
	output, err := GetModule(module)
	if err != nil {
		fmt.Println("ERROR")
		return Build{
			Module:  module,
			Success: false,
			Output:  string(output),
		}
	} else {
		defer os.RemoveAll(moduleDir)
		os.Chdir(moduleDir)
		// run the build command
		cmd := exec.Command("bash", "-c", module.Command)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("ERROR")
			return Build{
				Module:  module,
				Success: false,
				Output:  strings.TrimSpace(string(output)),
			}
		} else {
			fmt.Println("OK")
			return Build{
				Module:  module,
				Success: true,
				Output:  string(output),
			}
		}
	}
}

// BuildModules builds the list of modules in the configuration (in the exact same
// order).
func BuildModules(config Config) Builds {
	builds := make(Builds, len(config.Modules))
	repoStatus := LoadRepoHash(config.RepoStatus)
	for index, module := range config.Modules {
		if repoStatus[module.Name] == "" ||
			(repoStatus[module.Name] != GetRepoHash(module)) {
			builds[index] = BuildModule(module, config.Directory)
		}
	}
	return builds
}
