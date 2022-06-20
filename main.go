package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)


type Config struct {
	TARGET_HOST map[string]string `json:"target_host"`
	LOCAL_PORT string `json:"local_port"`
	ACCESS_CONTROL_ALLOWS_ORIGIN string `json:"access_control_allows_origin"`
	WITH_CREDENTIALS string `json:"with_credentials"`
	SESSION_COOKIE_NAME string `json:"session_cookie_name"`
	SSL_KEY_PATH string `json:"ssl_key_path"`
	SSL_CERT_PATH string `json:"ssl_cert_path"`
}

var (
	// hostTarget = map[string]string{
	// 	"user": "http://122.34.166.47:4101/",
	// 	"post": "http://localhost:4102/",
	// }
	hostProxy = make(map[string]*httputil.ReverseProxy)// = map[string]*httputil.ReverseProxy{}
	config Config
)

type baseHandle struct{}

func (h *baseHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	prefix := strings.Split(path, "/")[1]
	fmt.Println(path)

	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", config.ACCESS_CONTROL_ALLOWS_ORIGIN)
		w.Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Origin, content-type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Set-Cookie, Access-Control-Allow-Origin, Access-Control-Allow-Methods, Access-Control-Allow-Credential, Authorization")
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Vary", "Access-Control-Request-Method")
		w.Header().Set("Vary", "Access-Control-Request-Headers")
		w.Header().Set("Access-Control-Allow-Credentials", config.WITH_CREDENTIALS)
		return
	} else {
		log.Printf("%s %s -> GRP_Proxy -> %s", r.Method, r.Host+r.URL.RequestURI(), r.URL.RequestURI())
		if fn, ok := hostProxy[prefix]; ok {
			fn.ServeHTTP(w, r)
			return
		}
		if target, ok := config.TARGET_HOST[prefix]; ok {
			remoteUrl, err := url.Parse(target)
			if err != nil {
				log.Println("target parse fail:", err)
				return
			}
			log.Println(Yellow,"Forwarding Target : " + remoteUrl.Scheme + "://" + remoteUrl.Host + remoteUrl.Path, Reset)
	
			proxy := httputil.NewSingleHostReverseProxy(remoteUrl)
			proxy.ModifyResponse = corsHeaderModify
	
			hostProxy[prefix] = proxy
			proxy.ServeHTTP(w, r)
			return
		}
	}
	w.Write([]byte("Forbidden : " + prefix))
}

func main() {
	printConsoleMessage()
	readConfigFile();

	h := &baseHandle{}
	http.Handle("/", h)

	server := &http.Server{
		Addr:    ":" + config.LOCAL_PORT,
		Handler: h,
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	err := server.ListenAndServe()
	// err = http.ListenAndServeTLS(":"+os.Getenv("LOCAL_PORT"), os.Getenv("SSL_CERT_PATH"), os.Getenv("SSL_KEY_PATH"), nil)
	if err != nil {
		panic(err)
	}
}


func corsHeaderModify(resp *http.Response) error {
	// Set Basic Cors related header
	resp.Header.Set("Access-Control-Allow-Origin", config.ACCESS_CONTROL_ALLOWS_ORIGIN)
	resp.Header.Set("Access-Control-Allow-Headers", "Access-Control-Allow-Origin, content-type")
	resp.Header.Set("Access-Control-Allow-Methods", "*")
	resp.Header.Set("Access-Control-Expose-Headers", "Set-Cookie, Access-Control-Allow-Origin, Access-Control-Allow-Methods, Access-Control-Allow-Credential, Authorization")
	resp.Header.Set("Vary", "Origin")
	resp.Header.Set("Vary", "Access-Control-Request-Method")
	resp.Header.Set("Vary", "Access-Control-Request-Headers")
	resp.Header.Set("Access-Control-Allow-Credentials", config.WITH_CREDENTIALS)

	// Parsing cookie in header
	for _, value := range strings.Split(resp.Header.Get("Set-Cookie"), ";") {
		// If remove the domain value, the client host information is automatically set to the domain value by the browser.
		if strings.Contains(value, "Domain=") {
			var newCookie = strings.Replace(resp.Header.Get("Set-Cookie"), value, "", 1)
			resp.Header.Set("Set-Cookie", newCookie)
		}
	}
	return nil
}

func printConsoleMessage () {
	fmt.Println(Purple)
	fmt.Println(" ____     ____     ____           ____                                            ")
	fmt.Println("/\\  _`\\  /\\  _`\\  /\\  _`\\        /\\  _`\\                                          ")
	fmt.Println("\\ \\ \\L\\_\\\\ \\ \\L\\ \\\\ \\ \\L\\ \\      \\ \\,\\L\\_\\      __   _ __   __  __     __   _ __  ")
	fmt.Println(" \\ \\ \\L_L \\ \\ ,  / \\ \\ ,__/_______\\/_\\__ \\    /'__`\\/\\`'__\\/\\ \\/\\ \\  /'__`\\/\\`'__\\")
	fmt.Println("  \\ \\ \\/, \\\\ \\ \\\\ \\ \\ \\ \\//\\______\\ /\\ \\L\\ \\ /\\  __/\\ \\ \\/ \\ \\ \\_/ |/\\  __/\\ \\ \\/ ")
	fmt.Println("   \\ \\____/ \\ \\_\\ \\_\\\\ \\_\\\\/______/ \\ `\\____\\\\ \\____\\\\ \\_\\  \\ \\___/ \\ \\____\\\\ \\_\\ ")
	fmt.Println("    \\/___/   \\/_/\\/ / \\/_/           \\/_____/ \\/____/ \\/_/   \\/__/   \\/____/ \\/_/")
	fmt.Println(Reset)
}

func readConfigFile () {
	log.Println("Read config.json.")

	data, err := ioutil.ReadFile("./config.json")
	if err != nil {
		panic("Error: Cannot read config.json file.")
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		panic("Error: Cannot unmashal config.json to json object. (wrong syntax).")
	}
	log.Println("Success to read!")
}

var Reset  = "\033[0m"
var Red    = "\033[31m"
var Green  = "\033[32m"
var Yellow = "\033[33m"
var Purple = "\033[35m"
var White  = "\033[97m"