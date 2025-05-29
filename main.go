package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

var Logger *slog.Logger

type Event struct {
	*PushEvent
	*PackageEvent
}

type PackageEvent struct {
	Action  string  `json:"action"`
	Package Package `json:"package"`
}

type Package struct {
	CreatedAt      string         `json:"created_at"`
	Description    string         `json:"description"`
	Ecosystem      string         `json:"ecosystem"`
	HtmlUrl        string         `json:"html_url"`
	Id             int            `json:"id"`
	Name           string         `json:"name"`
	Namespace      string         `json:"namespace"`
	PackageType    string         `json:"package_type"`
	PackageVersion PackageVersion `json:"package_version"`
}

type PackageVersion struct {
	Name              string            `json:"name"`
	ContainerMetadata ContainerMetadata `json:"container_metadata"`
}

type ContainerMetadata struct {
	Tag Tag `json:"tag"`
}

type Tag struct {
	Name string `json:"name"`
}

type PushEvent struct {
	After   string   `json:"after"`
	Before  string   `json:"before"`
	Commits []Commit `json:"commits"`
}

type Commit struct {
	Id        string   `json:"id"`
	Added     []string `json:"added"`
	Modified  []string `json:"modified"`
	Removed   []string `json:"removed"`
	Message   string   `json:"message"`
	Timestamp string   `json:"timestamp"`
	TreeId    string   `json:"tree_id"`
	Url       string   `json:"url"`
}

func main() {
	Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	Logger.Info("Starting infrastructure service on port 8080")
	if err := http.ListenAndServe(":8080", http.HandlerFunc(Handle)); err != nil {
		panic(err)
	}
}

func Handle(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	logger := Logger.With("method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr)
	if r.URL.Path == "/_health" {
		w.Write([]byte("OK"))
		return
	}
	if r.URL.Path != "/" {
		logger.Error("Not found")
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodPost {
		logger.Error("Method not allowed", "method", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	repo := r.URL.Query().Get("repo")
	if repo == "" {
		logger.Error("Missing 'repo' query parameter")
		http.Error(w, "Missing 'repo' query parameter", http.StatusBadRequest)
		return
	}
	logger = logger.With("repo", repo)

	var e *Event
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&e); err != nil {
		logger.Error("Failed to decode request e", "error", err)
	}

	if e == nil {
		logger.Error("Request body is empty")
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}

	if e.PackageEvent != nil {
		logger.Info("package created",
			"action", e.Action,
			"package_name", e.Package.Name,
			"package_version", e.Package.PackageVersion.Name,
			"package_id", e.Package.Id,
			"package_created_at", e.Package.CreatedAt,
			"package_description", e.Package.Description,
			"package_ecosystem", e.Package.Ecosystem,
			"package_html_url", e.Package.HtmlUrl,
			"package_namespace", e.Package.Namespace,
			"package_type", e.Package.PackageType,
			"package_tag", e.Package.PackageVersion.ContainerMetadata.Tag.Name,
		)
		w.Write([]byte("Package created successfully"))
		return
	} else if e.PushEvent != nil {
		logger.Info("Push event received",
			"after", e.After,
			"before", e.Before,
			"commits_count", len(e.Commits),
			"commits", e.Commits,
		)
		w.Write([]byte("Push event handled successfully"))
	} else {
		logger.Error("Unsupported event")
		http.Error(w, "Unsupported event", http.StatusBadRequest)
		return
	}
}

type handler struct {
	Logger *slog.Logger
}

func (h *handler) handlePackage(w http.ResponseWriter, e *PackageEvent) {

}

func (h *handler) handlePush(w http.ResponseWriter, body map[string]interface{}) {
	// Placeholder for push handling logic
	h.Logger.Info("Push event received", "body", body)
	w.Write([]byte("Push event handled successfully"))
}
