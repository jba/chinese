package quiz

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

func init() {
	http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	switch r.URL.Path {
	case "/flashcards":
		runFlashcards(w, q.Get("n"))
	case "/quiz":
		runQuiz(w, q.Get("n"))
	case "/upload":
		handleUpload(w, r, q.Get("corpus"))
	default:
		http.Error(w, "use /flashcards or /quiz", http.StatusNotFound)
	}
}

func runFlashcards(w http.ResponseWriter, sn string) {
	n := 10
	if sn != "" {
		var err error
		n, err = strconv.Atoi(sn)
		if err != nil {
			http.Error(w, fmt.Sprintf("bad number: %v", err), http.StatusBadRequest)
			return
		}
	}
	fmt.Fprintf(w, "will do %d flashcards", n)
}

func runQuiz(w http.ResponseWriter, sn string) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func handleUpload(w http.ResponseWriter, r *http.Request, corpus string) {
	if corpus == "" {
		http.Error(w, "need corpus", http.StatusBadRequest)
		return
	}
	bytes, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		http.Error(w, fmt.Sprintf("bad body: %v", err), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "<html>Here's what I got:<pre>%s</pre></html>", string(bytes))
}
