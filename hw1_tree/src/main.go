package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	toFileMark   = "├───"
	lastFileMark = "└───"
	innerDirMark = "│"
	emptyMark    = "(empty)"
)

type fileInfoWithPath struct {
	FileInfo       os.FileInfo
	DrawParentMark string
	LastFile       bool
}

const pathSeparator = string(os.PathSeparator)

func fillPathsInfo(rootPath string, fullInfo bool) ([]string, map[string]*fileInfoWithPath, error) {
	var pathForProcessing []string
	fileMap := make(map[string]*fileInfoWithPath)
	handlePath := func(newpath, previousPath string) {
		if !strings.Contains(newpath, previousPath) {
			splitPath := strings.Split(newpath, pathSeparator)
			splitLastPath := strings.Split(previousPath, pathSeparator)
			var idx int
			var value string
			for idx, value = range splitPath {
				if splitLastPath[idx] != value {
					idx++
					break
				}
			}
			for idx < len(splitLastPath) {
				info := fileMap[filepath.Join(splitLastPath[:idx+1]...)]
				info.LastFile = true
				idx++
			}
		}
	}
	var lastPath string
	rootPathWithSeparator := rootPath + pathSeparator

	filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if rootPath == path || (!info.IsDir() && !fullInfo) {
			return nil
		}
		path = strings.Replace(path, rootPathWithSeparator, "", 1)
		handlePath(path, lastPath)
		lastPath = path
		pathForProcessing = append(pathForProcessing, path)
		fileMap[path] = &fileInfoWithPath{
			FileInfo: info,
		}
		return nil
	})
	splitLastPath := strings.Split(lastPath, pathSeparator)
	for idx := range splitLastPath {
		if path := filepath.Join(splitLastPath[:idx+1]...); rootPath != path {
			fileMap[path].LastFile = true
		}
	}
	return pathForProcessing, fileMap, nil
}

func drawPaths(out io.Writer, pathForProcessing []string, fileMap map[string]*fileInfoWithPath, fullInfo bool) {
	var previousMark []string
	for _, value := range pathForProcessing {
		previousMark = previousMark[:strings.Count(value, pathSeparator)]
		fileInfo := fileMap[value]
		var result string
		if len(previousMark) != 0 {
			result = fmt.Sprintf("%s\t", strings.Join(previousMark, "\t"))
		}
		if fileInfo.LastFile {
			result += lastFileMark + fileInfo.FileInfo.Name()
			previousMark = append(previousMark, "")
		} else {
			result += toFileMark + fileInfo.FileInfo.Name()
			previousMark = append(previousMark, innerDirMark)
		}
		if fullInfo && !fileInfo.FileInfo.IsDir() {
			switch size := fileInfo.FileInfo.Size(); size {
			case 0:
				result += fmt.Sprintf(" %s", emptyMark)
			default:
				result += fmt.Sprintf(" (%db)", size)
			}
		}
		fmt.Fprintln(out, result)
	}
}

func dirTree(out io.Writer, rootPath string, fullInfo bool) error {

	pathForProcessing, fileMap, _ := fillPathsInfo(rootPath, fullInfo)
	drawPaths(out, pathForProcessing, fileMap, fullInfo)
	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
