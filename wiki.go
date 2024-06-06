package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
)

// Types here
type Page struct {
	Title       string
	Body        []byte
	DisplayBody template.HTML
}

// Globals here
var templates = template.Must(template.ParseFiles("./tmpl/edit.html", "./tmpl/view.html"))
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// Functions here
func (p *Page) save() error {
	filename := "./data/" + p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

func createInterLink(s []byte) []byte {
	r, _ := regexp.Compile(`([a-zA-Z]+)`)
	l := r.FindString(string(s))
	h := ("<a href=\"/view/" + l + "\">" + l + "</a>")
	return []byte(h)
}

func loadPage(title string) (*Page, error) {
	filename := "./data/" + title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	regCom, _ := regexp.Compile(`\[([a-zA-Z]+)\]`)
	escapedBody := []byte(template.HTMLEscapeString(string(p.Body)))
	p.DisplayBody = template.HTML(regCom.ReplaceAllFunc(escapedBody, createInterLink))
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func homeRedirect(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
	return
}

// Main here
func main() {
	http.HandleFunc("/", homeRedirect)
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	log.Fatal(http.ListenAndServe(":8197", nil))
}
