package main

import (
	"log"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	yaml "gopkg.in/yaml.v2"
	"github.com/gorilla/handlers"
)

var configpath string

// MacNotFoundError error thrown when mac is not found in arp table
type MacNotFoundError struct {
	Msg string
}

func (e *MacNotFoundError) Error() string {
	return fmt.Sprintf(e.Msg)
}

func fileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}

func getConfig(r *http.Request) (map[interface{}]interface{}, error) {
	var data []byte
	var err error

	remoteaddr := r.Header.Get("X-Forwarded-For")
	if remoteaddr == "" {
		remoteaddr = strings.Split(r.RemoteAddr, ":")[0]
	}

	addressconfig := path.Join(configpath, remoteaddr)
	defaultconfig := path.Join(configpath, "default")

	if fileExists(addressconfig) {
		data, err = ioutil.ReadFile(addressconfig)
		if err != nil {
			return nil, err
		}
	} else { //fallback to default
		data, err = ioutil.ReadFile(defaultconfig)
		if err != nil {
			return nil, err
		}
	}

	var config map[interface{}]interface{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
			return nil, err
	}
	return config, nil
}

func metadata(w http.ResponseWriter, r *http.Request) {
	dirname, filename := path.Split(r.URL.Path)
	if dirname != "/latest/meta-data/" && dirname != "/2009-04-04/meta-data/" {
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
	metadata := config["meta-data"].(map[interface{}]interface{})
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
	userdata := config["user-data"].(map[interface{}]interface{})
	userdata["datasource"] = map[interface{}]map[string]bool{"Ec2": {"strict_id": false}}
	userdatabytes, err := yaml.Marshal(userdata)
	if err != nil {
		fmt.Println("Failed to get user-data metadata", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	data := "#cloud-config\n" + string(userdatabytes)
	w.Header().Set("Content-Type", "text/yaml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
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
	mux := http.NewServeMux()
	mux.HandleFunc("/2009-04-04/meta-data/", metadata)
	mux.HandleFunc("/latest/meta-data/", metadata)
	mux.HandleFunc("/2009-04-04/user-data", userData)
	mux.HandleFunc("/latest/user-data", userData)
	log.Fatal(http.ListenAndServe(bind, handlers.LoggingHandler(os.Stdout, mux)))
}
