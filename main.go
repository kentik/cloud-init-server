package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/mostlygeek/arp"

	yaml "gopkg.in/yaml.v2"
)

var configpath string

// MacNotFoundError error thrown when mac is not found in arp table
type MacNotFoundError struct {
	Msg string
}

func (e *MacNotFoundError) Error() string {
	return fmt.Sprintf(e.Msg)
}

func getConfig(r *http.Request) (map[string]interface{}, error) {
	remoteaddr := strings.Split(r.RemoteAddr, ":")[0]
	mac := arp.Search(remoteaddr)
	if mac == "" {
		return nil, &MacNotFoundError{Msg: fmt.Sprintf("Could not find mac for ip %s", remoteaddr)}
	}
	fullpath := path.Join(configpath, mac)
	data, err := ioutil.ReadFile(fullpath)
	if err != nil {
		return nil, err
	}
	var config map[string]interface{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
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
	config, err := getConfig(r)
	if err != nil {
		fmt.Println("Failed to get metadata", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	metadata := config["meta-data"].(map[string]interface{})
	w.WriteHeader(http.StatusOK)
	if filename == "" {
		fmt.Fprintln(w, "instance-id")
		for key := range metadata {
			fmt.Fprintln(w, key)
		}
	} else {
		fmt.Fprintf(w, metadata[filename].(string))
	}

}

func userData(w http.ResponseWriter, r *http.Request) {
	config, err := getConfig(r)
	if err != nil {
		fmt.Println("Failed to get user-data metadata", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	userdata := config["user-data"]
	userdatabytes, err := yaml.Marshal(userdata)
	if err != nil {
		fmt.Println("Failed to get user-data metadata", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "text/cloud-config")
	w.WriteHeader(http.StatusOK)
	w.Write(userdatabytes)
}

func main() {
	var bind string
	flag.StringVar(&bind, "bind", ":80", "Address to bind on defaults to :80")
	flag.StringVar(&configpath, "config", "/etc/cloud-init", "Path to put cloud-init config files in")
	flag.Parse()
	if _, err := os.Stat(configpath); os.IsNotExist(err) {
		fmt.Printf("Config path %s does not exists\n", configpath)
		os.Exit(1)
	}
	http.HandleFunc("/latest/meta-data/", metadata)
	http.HandleFunc("/latest/user-data", userData)
	http.ListenAndServe(bind, nil)
}
