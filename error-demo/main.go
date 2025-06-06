package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"
)

//
// ────────────────────────────────────────────────────────────
//   NOTE ON LOGGING:
//   • log.Println(...) → writes to stderr (Cloud Logging treats as ERROR).
//   • fmt.Println(...) → writes to stdout (Cloud Logging treats as INFO).
//   To produce a “WARNING,” write a single‐line JSON with "severity":"WARNING".
// ────────────────────────────────────────────────────────────
//

// homeHandler serves an HTML page with buttons to trigger different log severities.
func homeHandler(w http.ResponseWriter, r *http.Request) {
    // 1) Log an INFO to stdout so Cloud Logging marks it severity=INFO.
    fmt.Println("INFO:", "home page visited at", time.Now().Format(time.RFC3339))

    // 2) Serve a simple HTML page with four buttons:
    //    • /trigger-error
    //    • /trigger-panic
    //    • /trigger-warning
    //    • /trigger-custom
    fmt.Fprint(w, `
  <!DOCTYPE html>
  <html>
    <head><title>Error Demo</title></head>
    <body>
      <h1>Error Demo App</h1>
      <p>Click a button below to generate different log severities:</p>
      <ul>
        <li><form action="/trigger-error"><button>Trigger ERROR</button></form></li>
        <li><form action="/trigger-panic"><button>Trigger PANIC</button></form></li>
        <li><form action="/trigger-warning"><button>Trigger WARNING</button></form></li>
        <li><form action="/trigger-custom"><button>Trigger CUSTOM Error</button></form></li>
      </ul>
      <p>(Return to <a href="/">Home</a> to try again.)</p>
    </body>
  </html>`)
}

// errorHandler logs a basic ERROR and returns HTTP 500.
func errorHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("ERROR:", "generic error triggered by /trigger-error at", time.Now().Format(time.RFC3339))
    w.WriteHeader(http.StatusInternalServerError)
    fmt.Fprint(w, "500 Internal Server Error: generic error was triggered.\n")
}

// panicHandler logs an ERROR then panics (simulating a crash).
func panicHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("ERROR:", "about to panic (triggered by /trigger-panic) at", time.Now().Format(time.RFC3339))
    panic("✨ intentional panic: simulated crash for demo ✨")
}

// warningHandler emits a JSON “WARNING” log to stdout.
func warningHandler(w http.ResponseWriter, r *http.Request) {
    rec := map[string]interface{}{
        "severity": "WARNING",
        "message":  "This is a WARNING log triggered by /trigger-warning",
        "time":     time.Now().Format(time.RFC3339),
    }
    raw, _ := json.Marshal(rec)
    fmt.Println(string(raw)) // stdout → Cloud Logging picks up severity=WARNING

    w.WriteHeader(http.StatusOK)
    fmt.Fprint(w, "200 OK: a WARNING log was emitted.\n")
}

// customHandler emits a structured JSON “database error” to stderr.
func customHandler(w http.ResponseWriter, r *http.Request) {
    errRec := map[string]interface{}{
        "severity":     "ERROR",
        "errorType":    "DatabaseConnectionError",
        "description":  "Unable to connect to DB host 'db-primary:5432'",
        "retryable":    false,
        "timestamp":    time.Now().Format(time.RFC3339),
    }
    raw, _ := json.Marshal(errRec)
    fmt.Fprintln(os.Stderr, string(raw)) // stderr → severity=ERROR

    w.WriteHeader(http.StatusInternalServerError)
    fmt.Fprint(w, "500 Internal Server Error: database connection error simulated.\n")
}

func main() {
    // Log startup (INFO → stdout)
    fmt.Println("INFO:", "starting Error Demo server on port", getPort(), "at", time.Now().Format(time.RFC3339))

    // Register HTTP handlers
    http.HandleFunc("/", homeHandler)
    http.HandleFunc("/trigger-error", errorHandler)
    http.HandleFunc("/trigger-panic", panicHandler)
    http.HandleFunc("/trigger-warning", warningHandler)
    http.HandleFunc("/trigger-custom", customHandler)

    // Listen on the provided PORT (Cloud Run sets this); default to 8080 if unset.
    port := getPort()
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

// getPort returns the PORT env var or "8080" if not set.
func getPort() string {
    if p := os.Getenv("PORT"); p != "" {
        return p
    }
    return "8080"
}
