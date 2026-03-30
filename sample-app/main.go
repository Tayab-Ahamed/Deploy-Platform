package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = "hello-world"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <title>%s</title>
  <style>
    body { font-family: system-ui; background: #0d0d0f; color: #e8e8ec;
           display: flex; align-items: center; justify-content: center;
           min-height: 100vh; margin: 0; }
    .card { background: #141418; border: 1px solid rgba(255,255,255,0.1);
            border-radius: 16px; padding: 40px 48px; text-align: center; }
    h1 { font-size: 32px; margin: 0 0 8px; }
    p  { color: #7c7c8a; margin: 0; }
    .badge { display: inline-block; margin-top: 20px; padding: 4px 14px;
             background: rgba(34,199,122,0.12); color: #22c77a;
             border-radius: 20px; font-size: 13px; font-weight: 600; }
  </style>
</head>
<body>
  <div class="card">
    <h1>🚀 %s</h1>
    <p>Deployed via Deploy Platform</p>
    <div class="badge">● Live</div>
    <p style="margin-top:16px; font-size:12px; color:#444">%s</p>
  </div>
</body>
</html>`, appName, appName, time.Now().Format("2006-01-02 15:04:05 UTC"))
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Sample app '%s' listening on :%s", appName, port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
