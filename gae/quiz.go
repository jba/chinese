package quiz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/jba/chinese/study"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

func init() {
	http.HandleFunc("/", handler)
	rand.Seed(time.Now().UnixNano())
}

func handler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	switch r.URL.Path {
	case "/flashcards":
		handleStudy(appengine.NewContext(r), w, "flashcards", q.Get("n"), q.Get("corpus"))
	case "/quiz":
		handleStudy(appengine.NewContext(r), w, "quiz", q.Get("n"), q.Get("corpus"))
	case "/upload/items":
		handleUploadItems(w, r, q.Get("corpus"))
	case "/upload/words":
		handleUploadWords(w, r, q.Get("corpus"))
	case "/show":
		handleShow(w, r, q.Get("corpus"))
	case "/clear":
		handleDeleteCorpus(w, r, q.Get("corpus"))
	default:
		http.Error(w, "unknown path", http.StatusNotFound)
	}
}

func handleStudy(ctx context.Context, w http.ResponseWriter, kind, sn, corpusName string) {
	n, ok := parseInt(w, sn, 10)
	if !ok {
		return
	}
	if corpusName == "" {
		http.Error(w, "need corpus", http.StatusBadRequest)
		return
	}
	corpus, err := loadCorpus(ctx, corpusName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	entries := study.BuildEntries(corpus.Items, corpus.Words, n)
	bytes, err := json.MarshalIndent(entries, "", "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(bytes); err != nil {
		log.Println(err)
	}

}

func parseInt(w http.ResponseWriter, sn string, defaultValue int) (int, bool) {
	n := defaultValue
	if sn != "" {
		var err error
		n, err = strconv.Atoi(sn)
		if err != nil {
			http.Error(w, fmt.Sprintf("bad number: %v", err), http.StatusBadRequest)
			return 0, false
		}
	}
	return n, true
}

func runQuiz(w http.ResponseWriter, sn string) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func handleUploadItems(w http.ResponseWriter, r *http.Request, corpus string) {
	if corpus == "" {
		http.Error(w, "need corpus", http.StatusBadRequest)
		return
	}
	body, ok := readBody(w, r)
	if !ok {
		return
	}
	items, err := study.ParseItems(body)
	if err != nil {
		http.Error(w, fmt.Sprintf("parsing items: %v", err), http.StatusBadRequest)
		return
	}

	if err := saveItems(appengine.NewContext(r), items, corpus); err != nil {
		http.Error(w, fmt.Sprintf("saving items: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Got %d items.\n", len(items))
}

func handleUploadWords(w http.ResponseWriter, r *http.Request, corpus string) {
	if corpus == "" {
		http.Error(w, "need corpus", http.StatusBadRequest)
		return
	}
	body, ok := readBody(w, r)
	if !ok {
		return
	}
	words, err := study.ParseWords(body)
	if err != nil {
		http.Error(w, fmt.Sprintf("parsing words: %v", err), http.StatusBadRequest)
		return
	}

	if err := saveWords(appengine.NewContext(r), words, corpus); err != nil {
		http.Error(w, fmt.Sprintf("saving words: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Got %d words.\n", len(words))
}

func readBody(w http.ResponseWriter, r *http.Request) (string, bool) {
	bytes, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		http.Error(w, fmt.Sprintf("bad body: %v", err), http.StatusBadRequest)
		return "", false
	}
	return string(bytes), true
}

var showCorpusTemplate = template.Must(template.New("").Parse(`
<html>
  <body>
    <h1>{{.Name}}</h1>
    <h2>Items</h2>
    <table>
	  {{range .Items}}
		 <tr><td>{{.English}}</td><td>{{.Pinyin}}</td><td>{{.Characters}}</td></tr>
	  {{else}}
		 Nothing there.
	  {{end}}
    </table>
    <h2>Words</h2>
    <table>
	  {{range .Words}}
		 <tr><td>{{.English}}</td><td>{{.Pinyin}}</td><td>{{.PartOfSpeech}}</td><td>{{.Characters}}</td></tr>
	  {{else}}
		 Nothing there.
	  {{end}}
    </table>
  </body>
</html>
`))

func handleShow(w http.ResponseWriter, r *http.Request, corpus string) {
	if corpus == "" {
		http.Error(w, "need corpus", http.StatusBadRequest)
		return
	}
	items, err := loadItems(appengine.NewContext(r), corpus)
	if err != nil {
		http.Error(w, fmt.Sprintf("loading items: %v", err), http.StatusInternalServerError)
		return
	}
	words, err := loadWords(appengine.NewContext(r), corpus)
	if err != nil {
		http.Error(w, fmt.Sprintf("loading words: %v", err), http.StatusInternalServerError)
		return
	}
	corp := Corpus{
		Name:  corpus,
		Items: items,
		Words: words,
	}
	var buf bytes.Buffer
	if err := showCorpusTemplate.Execute(&buf, corp); err != nil {
		http.Error(w, fmt.Sprintf("executing template: %v", err), http.StatusInternalServerError)
		return
	}
	if _, err := buf.WriteTo(w); err != nil {
		log.Println(err)
	}
}

func handleDeleteThing(w http.ResponseWriter, r *http.Request, corpus, key string) {
	if corpus == "" {
		http.Error(w, "need corpus", http.StatusBadRequest)
		return
	}
	if key == "" {
		http.Error(w, "need key", http.StatusBadRequest)
		return
	}
	ctx := appengine.NewContext(r)
	err := datastore.Delete(ctx, datastore.NewKey(ctx, "Item", key, 0, corpusKey(ctx, corpus)))
	if err != nil {
		errorf(w, http.StatusInternalServerError, "deleting item: %v", err)
		return
	}
	err = datastore.Delete(ctx, datastore.NewKey(ctx, "Word", key, 0, corpusKey(ctx, corpus)))
	if err != nil {
		errorf(w, http.StatusInternalServerError, "deleting word: %v", err)
		return
	}
}

func handleDeleteCorpus(w http.ResponseWriter, r *http.Request, corpus string) {
	if corpus == "" {
		errorf(w, http.StatusBadRequest, "need corpus")
		return
	}
	ctx := appengine.NewContext(r)
	q := datastore.NewQuery("Item").Ancestor(corpusKey(ctx, corpus)).KeysOnly()
	keys, err := q.GetAll(ctx, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("reading keys: %v", err), http.StatusInternalServerError)
		return
	}
	if err := datastore.DeleteMulti(ctx, keys); err != nil {
		errorf(w, http.StatusInternalServerError, "deleting: %v", err)
		return
	}
	fmt.Fprintf(w, "Deleted %d keys.\n", len(keys))
}

func errorf(w http.ResponseWriter, status int, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	http.Error(w, msg, status)
	log.Print(msg)
}

////////////////////////////////////////////////////////////////
// Internals.

type Corpus struct {
	Name  string
	Items []*study.Item
	Words []*study.Word
}

func corpusKey(ctx context.Context, corpus string) *datastore.Key {
	return datastore.NewKey(ctx, "Corpus", corpus, 0, nil)
}

func saveItems(ctx context.Context, items []*study.Item, corpus string) error {
	for _, it := range items {
		key := datastore.NewKey(ctx, "Item", it.English, 0, corpusKey(ctx, corpus))
		_, err := datastore.Put(ctx, key, it)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadItems(ctx context.Context, corpus string) ([]*study.Item, error) {
	q := datastore.NewQuery("Item").Ancestor(corpusKey(ctx, corpus))
	var items []*study.Item
	if _, err := q.GetAll(ctx, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func saveWords(ctx context.Context, words []*study.Word, corpus string) error {
	for _, w := range words {
		key := datastore.NewKey(ctx, "Word", w.English, 0, corpusKey(ctx, corpus))
		_, err := datastore.Put(ctx, key, w)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadWords(ctx context.Context, corpus string) ([]*study.Word, error) {
	q := datastore.NewQuery("Word").Ancestor(corpusKey(ctx, corpus))
	var words []*study.Word
	if _, err := q.GetAll(ctx, &words); err != nil {
		return nil, err
	}
	return words, nil
}

func loadCorpus(ctx context.Context, name string) (*Corpus, error) {
	items, err := loadItems(ctx, name)
	if err != nil {
		return nil, err
	}
	words, err := loadWords(ctx, name)
	if err != nil {
		return nil, err
	}
	return &Corpus{
		Name:  name,
		Items: items,
		Words: words,
	}, nil
}
