package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	mux := http.NewServeMux()
	hosturl := "127.0.0.1"
	hostport := "8080"
	password := "friendsonly"
	if len(os.Args) > 2 {
		hosturl = "afnetl.ink"
		hostport = "443"
		password = os.Args[1]
	}

	mux.HandleFunc(`/file`, func(w http.ResponseWriter, r *http.Request) {
		res := strings.Split(r.RequestURI, "=")
		if len(res) < 2 {
			return
		}
		uuid := res[1]
		dirEntrys, _ := os.ReadDir("database/" + uuid)
		fn := "database/" + uuid + "/" + dirEntrys[0].Name()
		w.Header().Add("Content-Disposition", "attachment; filename="+dirEntrys[0].Name())
		http.ServeFile(w, r, fn)
	})

	mux.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {

		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			log.Printf("Error parsing form: %s\n", err)
			return
		}

		values, ok := r.MultipartForm.File["filename"]
		if !ok {
			log.Printf("No multipart form filename key included")
			return
		}

		pwd, ok := r.MultipartForm.Value["pwd"]
		if !ok {
			log.Println("no pwd")
			return
		}
		if string(pwd[0]) != password {
			log.Printf("wrong password %s", pwd[0])
			w.Write([]byte("Incorrect password."))
			return
		}

		if len(values) < 1 {
			log.Printf("No filenames included")
			return
		}
		fn := values[0].Filename

		file, err := values[0].Open()
		if err != nil {
			log.Println("failed to parse file")
			return
		}

		rawFile, err := io.ReadAll(file)
		if err != nil {
			log.Println("Failed to read file")
			return
		}

		checksum := fmt.Sprintf("%x", sha256.Sum256(rawFile))
		uuid := checksum[:8]
		fmt.Printf("checksum: %s uuid: %s ", checksum, uuid)
		parentDirectory := "database/" + uuid

		if err := os.MkdirAll(parentDirectory, 0777); err != nil {
			log.Println(err)
			return
		}

		filepath := parentDirectory + "/" + fn
		if err := os.WriteFile(filepath, rawFile, 0777); err != nil {
			log.Println(err)
			return
		}

		w.Write([]byte(hosturl + "/file?id=" + uuid))

	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		f := http.FileServer(http.Dir("src"))
		f.ServeHTTP(w, r)
	})

	if len(os.Args) > 1 {
		log.Fatal(http.ListenAndServeTLS(hosturl+":"+hostport, "/etc/letsencrypt/live/afnetl.ink/fullchain.pem", "/etc/letsencrypt/live/afnetl.ink/privkey.pem", mux))
	}

	log.Fatal(http.ListenAndServe(hosturl+":"+hostport, mux))
}
