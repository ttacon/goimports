// determine_imports.go
package main

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	toAddToCommon = make(map[string]string)
	currentDir    string
	goPath        = os.Getenv("GOPATH")
)

func exportable(name string) bool {
	return len(name) > 0 && name[0] > 'A' && name[0] < 'Z'
}

type exportsToPackages struct {
	exports  []string
	packages []string
}

func GetImportsFromGoPath(regenIndex bool) map[string]string {
	// fmt.Fprintf(os.Stderr, "\"%v\"\n", os.Getenv("GOPATH"))
	fileName := filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "ttacon", "goimports", "goPathImports.json")
	fileInfo, err := os.Stat(fileName)
	// fmt.Fprintf(os.Stderr, "%v\n", err)
	if err != nil || regenIndex || time.Now().Sub(fileInfo.ModTime()).Hours() > 24.0 {
		// it either doesn't exist or it's unuseable, so let's make a new one
		filepath.Walk(filepath.Join(goPath, "src/"), func(path string, info os.FileInfo, err error) error {
			fset := token.NewFileSet()
			currentDir = path[len(filepath.Join(goPath, "src/")):]
			if info.IsDir() {
				mappings, err := parser.ParseDir(fset, path, isGoFile, 0)
				for _, v := range mappings {
					for _, v2 := range v.Files {
						for _, v3 := range v2.Scope.Objects {
							if exportable(v3.Name) && len(currentDir) > 0 && !strings.Contains(currentDir, "code.google.com/p/go") {
								toAddToCommon[v.Name+"."+v3.Name] = currentDir[1:]
							}
						}
					}
				}
				if err != nil {
					fmt.Println(err)
					return nil
				}
			}
			return nil
		})
		file, err := os.Create(fileName)
		if err != nil {
			fmt.Fprintln(os.Stderr, "unable to create file to save determined GOPATH imports")
			return toAddToCommon
		}
		// write to file for next time so we don't have to create this index
		fmt.Fprintf(file, "{\n")
		numKeysSeen := 0
		for selector, importPath := range toAddToCommon {
			numKeysSeen++
			lastComma := ","
			if numKeysSeen == len(toAddToCommon) {
				lastComma = ""
			}
			fmt.Fprintf(file, "    %-46s%s\n", "\""+selector+"\":", "\""+importPath+"\""+lastComma)
		}
		fmt.Fprintf(file, "}\n")
		err = file.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to close file...")
		}
		return toAddToCommon
	} else {
		// the file exists with all this data already and it's not too old so it should already exist
		file, err := os.Open(fileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "died trying to open file\n")
			os.Exit(0) // replace all of these with returning make(map[string]string)
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "died trying to read everything from the file\n")
			os.Exit(0)
		}
		goPathDeterminedImports := make(map[string]string)
		err = json.Unmarshal(data, &goPathDeterminedImports)
		if err != nil {
			fmt.Fprintf(os.Stderr, "died trying to umarshal the json\n")
			fmt.Fprintln(os.Stderr, err)
			os.Exit(0)
		}
		if len(goPathDeterminedImports) == 0 {
			return GetImportsFromGoPath(true)
		}
		return goPathDeterminedImports
	}
}
