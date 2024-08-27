// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

package main

import (
	"html/template"
	"regexp"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Page struct {
	Title string
    Body  []byte // []byte is expected by i/o libraries in Go 
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return os.WriteFile(filename, p.Body, 0600)
}

// func loadPage(title string) *Page {
//     filename := title + ".txt"
//     body, _ := os.ReadFile(filename)
//     return &Page{Title: title, Body: body}
// }

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Here we will extract the page title from the Request,
        // and call the provided handler 'fn'
		m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return
        }
        fn(w, r, m[2])
    }
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
        http.NotFound(w, r)
        return "", errors.New("invalid Page Title")
    }
    return m[2], nil // The title is the second subexpression.
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    // t, err := template.ParseFiles(tmpl + ".html")
    // if err != nil {
    //     http.Error(w, err.Error(), http.StatusInternalServerError)
    //     return
    // }
    // err = t.Execute(w, p)

	err := templates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	// title := r.URL.Path[len("/view/"):]

	// title, err := getTitle(w, r)
    // if err != nil {
    //     return
    // }

	p, err := loadPage(title)
	if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
	renderTemplate(w, "view", p)

	// fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.Title, p.Body)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    // title := r.URL.Path[len("/edit/"):]

	// title, err := getTitle(w, r)
    // if err != nil {
    //     return
    // }

    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)

	// fmt.Fprintf(w, "<h1>Editing %s</h1>"+
	// 	"<form action=\"/save/%s\" method=\"POST\">"+
	// 	"<textarea name=\"body\">%s</textarea><br>"+
	// 	"<input type=\"submit\" value=\"Save\">"+
	// 	"</form>",
	// 	p.Title, p.Title, p.Body)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    // title := r.URL.Path[len("/save/"):]

	// title, err := getTitle(w, r)
    // if err != nil {
    //     return
    // }

    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body)}
    err := p.save()
	if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func main() {
	// p1 := &Page{Title: "TestPage", Body: []byte("This is a sample Page.")}
    // p1.save()
    // p2, _ := loadPage("TestPage")
    // fmt.Println(string(p2.Body))

    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

