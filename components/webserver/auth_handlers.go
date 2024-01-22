package webserver

import (
	"HighFrequencyDNSChecker/components/logger"
	"HighFrequencyDNSChecker/components/tools"
	"html/template"
	"net/http"

	"github.com/sirupsen/logrus"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("login.html").ParseFS(tmplFS, "html/login.html", "html/auth_styles.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if r.Method == "GET" {
		err := tmpl.Execute(w, nil) // Pass any data needed for template
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        
    } else if r.Method == "POST" {
        r.ParseForm()
        username := r.FormValue("username")
        password := r.FormValue("password")

        if !CheckCredentials(username, password) {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            logger.Audit.WithFields(logrus.Fields{
                "username": username,
                "token": "",
                "ip":       tools.GetClientIP(r),
                "status":   "failed",
            }).Warn("Authentication attempt failed")
            return
        }

        tokenString, err := GenerateToken(username)
        if err != nil {
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            logger.Audit.WithFields(logrus.Fields{
                "username": username,
                "token": "",
                "ip":       tools.GetClientIP(r),
                "status":   "failed",
            }).Error("Error generating session token")
            return
        }

        http.SetCookie(w, &http.Cookie{
            Name:   "session_token",
            Value:  tokenString,
            Path:   "/",
            MaxAge: 300,
        })

        logger.Audit.WithFields(logrus.Fields{
            "username": username,
            "token": tools.TruncateToken(tokenString),
            "ip":       tools.GetClientIP(r),
            "status":   "success",
        }).Info("Authentication success")

        http.Redirect(w, r, "/", http.StatusFound)
    }
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("session_token")
    if err != nil {
        // Handle error - cookie not found
        logger.Audit.WithFields(logrus.Fields{
            "action": "logout_attempt",
            "status": "failed",
            "reason": "session_token not found",
        }).Warn("Logout attempt failed")
        // Redirect to login or handle otherwise
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    // Log the token (consider truncating or hashing for security)
    tokenString := tools.TruncateToken(cookie.Value)

    logger.Audit.WithFields(logrus.Fields{
        "action": "logout",
        "token":  tokenString,
    }).Info("User logged out")

    http.SetCookie(w, &http.Cookie{
        Name:   "session_token",
        Value:  "",
        Path:   "/",
        MaxAge: -1,
    })
    http.Redirect(w, r, "/login", http.StatusFound)
}