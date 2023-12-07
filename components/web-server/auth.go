package webserver

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"github.com/nir0k/HighFrequencyDNSChecker/components/log"
	"github.com/sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("your_secret_key")

type Claims struct {
    Username string `json:"username"`
    jwt.StandardClaims
}


func GenerateToken(username string) (string, error) {
    expirationTime := time.Now().Add(5 * time.Minute)
    claims := &Claims{
        Username: username,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtKey)

    return tokenString, err
}


func ValidateToken(tokenString string) (*jwt.Token, error) {
    claims := &Claims{}

    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        return jwtKey, nil
    })

    return token, err
}


func loginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        // Display login form
        fmt.Fprintf(w, `
        <html>
        <head>
            <style>
                body {
                    margin: 0;
                    padding: 0;
                    height: 100vh;
                    display: flex;
                    justify-content: center;
                    align-items: center;
                    background-color: #f0f0f0;
                }
                form {
                    padding: 20px;
                    border: 1px solid #ddd;
                    border-radius: 5px;
                    background-color: #fff;
                }
            </style>
        </head>
        <body>
            <form action="/login" method="post">
                <div>
                    <label for="username">Username:</label><br>
                    <input type="text" id="username" name="username">
                </div>
                <div>
                    <label for="password">Password:</label><br>
                    <input type="password" id="password" name="password">
                </div>
                <div>
                    <input type="submit" value="Login">
                </div>
            </form>
        </body>
        </html>
        `)
    } else if r.Method == "POST" {
        r.ParseForm()
        username := r.FormValue("username")
        password := r.FormValue("password")

        if !CheckCredentials(username, password) {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            log.AuthLog.WithFields(logrus.Fields{
                "username": username,
                "token": "",
                "ip":       GetClientIP(r),
                "status":   "failed",
            }).Warn("Authentication attempt failed")
            return
        }

        tokenString, err := GenerateToken(username)
        if err != nil {
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            log.AuthLog.WithFields(logrus.Fields{
                "username": username,
                "token": "",
                "ip":       GetClientIP(r),
                "status":   "success",
            }).Error("Error generating session token")
            return
        }

        http.SetCookie(w, &http.Cookie{
            Name:   "session_token",
            Value:  tokenString,
            Path:   "/",
            MaxAge: 300,
        })

        log.AuthLog.WithFields(logrus.Fields{
            "username": username,
            "token": TruncateToken(tokenString),
            "ip":       GetClientIP(r),
            "status":   "success",
        }).Info("Authentication success")

        http.Redirect(w, r, "/", http.StatusFound)
    }
}


func authMiddleware(handler http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("session_token")
        if err != nil {
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }

        tokenString := cookie.Value
        token, err := ValidateToken(tokenString)
        if err != nil || !token.Valid {
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }

        // Optionally, refresh the token if it's close to expiration
        if shouldRefreshToken(token) {
            newTokenString, err := GenerateToken(getUsernameFromToken(token))
            if err != nil {
                log.AuthLog.WithFields(logrus.Fields{
                    "old_token": TruncateToken(tokenString),
                    "new_token": TruncateToken(newTokenString),
                }).Info("JWT token refreshed")
                http.Redirect(w, r, "/login", http.StatusFound)
                return
            }

            setSessionCookie(w, newTokenString)
        }

        handler(w, r)
    }
}


func shouldRefreshToken(token *jwt.Token) bool {
    // Define your logic to decide when to refresh the token
    // For example, if the token expires in less than 5 minute
    const refreshInterval = 5 * time.Minute
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return time.Until(time.Unix(claims.ExpiresAt, 0)) < refreshInterval
    }
    return false
}


func getUsernameFromToken(token *jwt.Token) string {
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims.Username
    }
    return ""
}

func setSessionCookie(w http.ResponseWriter, token string) {
    http.SetCookie(w, &http.Cookie{
        Name:   "session_token",
        Value:  token,
        Path:   "/",
        MaxAge: 300, // Adjust as per requirement
    })
}


func logoutHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("session_token")
    if err != nil {
        // Handle error - cookie not found
        log.AuthLog.WithFields(logrus.Fields{
            "action": "logout_attempt",
            "status": "failed",
            "reason": "session_token not found",
        }).Warn("Logout attempt failed")
        // Redirect to login or handle otherwise
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    // Log the token (consider truncating or hashing for security)
    tokenString := TruncateToken(cookie.Value) // Implement truncateToken as needed

    log.AuthLog.WithFields(logrus.Fields{
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


func CheckCredentials(username, password string) bool {
    envContent, err := os.ReadFile(".env")
    if err != nil {
        return false
    }

    lines := strings.Split(string(envContent), "\n")
    var envUsername, envPassword string
    for _, line := range lines {
        parts := strings.SplitN(line, "=", 2)
        if len(parts) == 2 {
            if parts[0] == "WATCHER_WEB_USER" {
                envUsername = parts[1]
            } else if parts[0] == "WATCHER_WEB_PASSWORD" {
                envPassword = parts[1]
            }
        }
    }
    return username == envUsername && password == envPassword
}
