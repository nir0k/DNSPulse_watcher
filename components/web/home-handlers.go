package webserver

import (
	"html/template"
	"net/http"
)


func homeHandler(w http.ResponseWriter, r *http.Request) {

    tmpl, err := template.ParseFS(tmplFS, "html/home.html", "html/navbar.html", "html/styles.html")
    if err != nil {
        http.Error(w, "Unable to parse template", http.StatusInternalServerError)
        return
    }
    err =  tmpl.Execute(w, nil)
    if err != nil {
        http.Error(w, "Failed to execute template: "+err.Error(), http.StatusInternalServerError)
        return
    }
}