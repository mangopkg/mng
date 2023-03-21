package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/urfave/cli/v2"
)

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func genHandlerService(rname string) {

	ls := capitalizeFirst(rname)

	s1 := fmt.Sprintf(`
package %[2]s

import (
	"net/http"
	"reflect"

	"github.com/mangopkg/mango"
)

type %[1]sHandler struct {
	Response    mango.Response
	MountAt     string
	%[1]sService   %[1]sService
	Methods     map[string]func(http.ResponseWriter, *http.Request)
}

func NewHandler(service %[1]sService) {

	h := %[1]sHandler{
		MountAt:     "/%[2]s",
		%[1]sService:   service,
		Methods:     make(map[string]func(http.ResponseWriter, *http.Request)),
	}

	f := reflect.TypeOf(&h)
	v := reflect.ValueOf(&h)

	h.%[1]sService.SetupHandler(h.MountAt, f, v, h.Methods)
}

/*
<@route{
"pattern": "/find",
"func": "Find",
"method": "GET"
}>
*/
func (h *%[1]sHandler) Find() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		h.Response.Message = "Successful"
		h.Response.StatusCode = 200
		h.Response.Data = h.%[1]sService.Get%[1]s()
		h.Response.Send(w)
	}
}
	`, ls, rname)

	s2 := fmt.Sprintf(`
package %[2]s

import "github.com/mangopkg/mango"

type %[1]sService struct {
	mango.Service
}

func NewService(s mango.Service) {
	nS := %[1]sService{
		s,
	}

	NewHandler(nS)
}

func (s *%[1]sService) Get%[1]s() string {
	return "hello world %s!"
}
	`, ls, rname)

	err := os.Mkdir(rname, 0755)
	if err != nil {
		log.Fatal(err)
	}

	err2 := os.WriteFile(rname+"/handler.go", []byte(s1), 0644)

	if err2 != nil {
		log.Fatal(err2)
	}

	err3 := os.WriteFile(rname+"/service.go", []byte(s2), 0644)
	if err3 != nil {
		log.Fatal(err3)
	}

	fmt.Println("Created route:", rname)
}

func isAlphabetical(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func isAlphaOrHyphen(s string) bool {
	re := regexp.MustCompile(`^[a-zA-Z-]+$`)
	return re.MatchString(s)
}

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "add a route",
				Action: func(cCtx *cli.Context) error {
					rname := cCtx.Args().First()
					ok := isAlphabetical(rname)
					if ok {
						genHandlerService(rname)
					} else {
						fmt.Println("ERROR: Invalid name")
					}
					return nil
				},
			},
			{
				Name:    "new",
				Aliases: []string{"n"},
				Usage:   "create a new mango app",
				Action: func(cCtx *cli.Context) error {
					rname := cCtx.Args().First()
					ok := isAlphaOrHyphen(rname)
					if ok {
						downloadAndExtractZip("https://github.com/mangopkg/create-mango-app/archive/refs/heads/main.zip", "./"+rname, true, rname)
					} else {
						fmt.Println("ERROR: Invalid name, it can only be alphabetical and can only contain -")
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func downloadAndExtractZip(url string, folderName string, skipRootFolder bool, name string) {

	err := os.MkdirAll(folderName, 0755)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	reader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, f := range reader.File {
		rc, err := f.Open()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer rc.Close()

		path := filepath.Join(folderName, f.Name)

		if skipRootFolder {
			pathParts := strings.Split(f.Name, "/")
			if len(pathParts) > 1 {
				path = filepath.Join(folderName, strings.Join(pathParts[1:], "/"))
			} else {
				continue
			}
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(path,
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
				f.Mode())
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}
	}

	replaceLineInFile(folderName+"/go.mod", 1, "module "+name)
	replaceLineInFile(folderName+"/api/api.go", 9, "\""+name+"/book\"")
	replaceLineInFile(folderName+"/main.go", 3, "import \""+name+"/api\"")

}

func replaceLineInFile(filePath string, lineNumber int, newText string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if lineNumber < 0 || lineNumber >= len(lines) {
		fmt.Println("line number out of range")
		return
	}

	lines[lineNumber-1] = newText

	output := strings.Join(lines, "\n")
	erros := os.WriteFile(filePath, []byte(output), 0644)
	if erros != nil {
		fmt.Println(erros.Error())
		return
	}
}
