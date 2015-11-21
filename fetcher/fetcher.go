package fetcher

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

// Serve will start the fetcher server and block until it stops. Since it blocks, it's a best practice to execute this func in a goroutine.
func Serve(port int) {
	rtr := mux.NewRouter()
	rtr.HandleFunc("/git/home/{name}/tar", getTar).Methods("GET")
	rtr.HandleFunc("/git/home/{name}/slug", getSlug).Methods("GET")
	rtr.HandleFunc("/git/home/health", health).Methods("GET")
	rtr.HandleFunc("/git/home/{name}/push", putSlug).Methods("PUT")
	hostStr := fmt.Sprintf(":%d", port)
	http.ListenAndServe(hostStr, rtr)
}

func getTar(w http.ResponseWriter, r *http.Request) {
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
	params := mux.Vars(r)
	name := params["name"]
	dat, err := ioutil.ReadFile(slugdirectory + name + "/slug.tgz")
	if err != nil {
		w.Write([]byte(name + "dosn't exist"))
	}
	w.Write(dat)
}

func putSlug(w http.ResponseWriter, r *http.Request) {
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
