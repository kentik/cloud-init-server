package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
)

func getConfig(r *http.Request, filename string) ([]byte, error) {
	remoteaddr := strings.Split(r.RemoteAddr, ":")[0]
	fullpath := path.Join(remoteaddr, filename)
	return ioutil.ReadFile(fullpath)
}

func metadata(w http.ResponseWriter, r *http.Request) {
	dirname, filename := path.Split(r.URL.Path)
	if dirname != "/latest/meta-data/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if filename == "instance-id" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "iid-datasource-cloudstack")
		return

	}
	config, err := getConfig(r, "meta-data")
	if err != nil {
		fmt.Println("Failed to get metadata", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var metadata map[string]string
	err = json.Unmarshal(config, &metadata)
	if err != nil {
		fmt.Println("Failed to unmarshal metadata", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	if filename == "" {
		fmt.Fprintln(w, "instance-id")
		for key := range metadata {
			fmt.Fprintln(w, key)
		}
	} else {
		fmt.Fprintf(w, metadata[filename])
	}

}

func userData(w http.ResponseWriter, r *http.Request) {
	config, err := getConfig(r, "user-data")
	if err != nil {
		fmt.Println("Failed to get user-data metadata", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/cloud-config")
	w.WriteHeader(http.StatusOK)
	w.Write(config)
}

func main() {
	http.HandleFunc("/latest/meta-data/", metadata)
	http.HandleFunc("/latest/user-data", userData)
	http.ListenAndServe(":80", nil)
}
