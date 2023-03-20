package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"

	"github.com/urfave/cli/v2"
)

func genHandlerService(rname string) {

	ls := strings.Title(rname)

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
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
