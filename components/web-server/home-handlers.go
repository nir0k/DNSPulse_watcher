package webserver

import (
	"fmt"
	"net/http"
)


func homeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, `
    <html>
    <head>
        <title>Config Editor</title>
        <style>
            .navbar { background-color: #333; overflow: hidden; }
            .navbar a { float: left; display: block; color: white; text-align: center; padding: 14px 16px; text-decoration: none; }
            .navbar a.logout { float: right; }
        </style>
    </head>
    <body>
        <div class="navbar">
            <a href="/">Home</a>
            <a href="/csv">Edit CSV File</a>
            <a href="/env">Show/Edit Configuration</a>
            <a href="/logout" class="logout">Logout</a>
        </div>
        <h1>Welcome to the Config Editor</h1>
    </body>
    </html>
    `)
}