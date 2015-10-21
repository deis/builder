package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

const (
	appdirectory  = "/home/git/"
	slugdirectory = "/apps/"
	cmdstring     = "/tmp/builder/build.sh"
)

func main() {
	rtr := mux.NewRouter()
	rtr.HandleFunc("/git/home/{name}/tar", getTar).Methods("GET")
	rtr.HandleFunc("/git/home/{name}/slug", getSlug).Methods("GET")
	rtr.HandleFunc("/git/home/health", health).Methods("GET")
	rtr.HandleFunc("/git/home/{name}/push", putSlug).Methods("PUT")
	http.Handle("/", rtr)
	log.Println("Listening... on 3000")
	http.ListenAndServe(":3000", nil)
}

func getTar(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hello, world! gettar")
	params := mux.Vars(r)
	name := strings.Split(params["name"], ":")[0]
	dat, err := ioutil.ReadFile(appdirectory + name + ".git/" + name + ".tar.gz")
	if err != nil {
		w.Write([]byte(name + "dosn't exist"))
	}
	w.Write(dat)
}

func health(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, world!")
}

func getSlug(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hello, world! getslug")
	params := mux.Vars(r)
	name := params["name"]
	dat, err := ioutil.ReadFile(slugdirectory + name + "/slug.tgz")
	if err != nil {
		w.Write([]byte(name + "dosn't exist"))
	}
	w.Write(dat)
}

func putSlug(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hello, world! putslug")
	params := mux.Vars(r)
	name := params["name"]
	log.Println(name)
	err := os.MkdirAll(slugdirectory+name, 0755)
	if err != nil {
		fmt.Println(err)
	}
	output, err := os.Create(slugdirectory + name + "/slug.tgz")
	if err != nil {
		fmt.Println(err)
	}
	defer output.Close()
	defer r.Body.Close()
	fmt.Println(r.ContentLength)
	io.Copy(output, r.Body)
	return
}
