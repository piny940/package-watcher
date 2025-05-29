package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

var Logger *slog.Logger

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

	var body map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&body); err != nil {
		logger.Error("Failed to decode request body", "error", err)
	}

	if body == nil {
		logger.Error("Request body is empty")
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}

	h := &handler{Logger: logger}
	if _, ok := body["package"]; ok {
		h.handlePackage(w, body)
		return
	} else {
		logger.Error("Unsupported event")
		http.Error(w, "Unsupported event", http.StatusBadRequest)
		return
	}
}

type handler struct {
	Logger *slog.Logger
}

func (h *handler) handlePackage(w http.ResponseWriter, body map[string]interface{}) {
	p, ok := body["package"].(map[string]interface{})
	if !ok {
		h.Logger.Error("Package field is missing or not an object")
		http.Error(w, "Package field is missing or not an object", http.StatusBadRequest)
		return
	}
	args := toArgs(p, "")
	reg, ok := p["registry"].(map[string]interface{})
	if !ok {
		h.Logger.Error("Registry field is missing or not an object")
		http.Error(w, "Registry field is missing or not an object", http.StatusBadRequest)
		return
	}
	args = append(args, toArgs(reg, "registry.")...)
	v, ok := p["package_version"].(map[string]interface{})
	if !ok {
		h.Logger.Error("Package version field is missing or not an object")
		http.Error(w, "Package version field is missing or not an object", http.StatusBadRequest)
		return
	}
	args = append(args, toArgs(v, "package_version.")...)
	args = append(args, "tag", v["container_metadata"].(map[string]interface{})["tag"].(map[string]interface{})["name"].(string))

	argsAny := make([]any, len(args))
	for i, v := range args {
		argsAny[i] = v
	}
	h.Logger.Info("package created", argsAny...)
	w.Write([]byte("Package created successfully"))
}

func toArgs(body map[string]interface{}, prefix string) []string {
	args := make([]string, 0, len(body)*2)
	for key, value := range body {
		if strValue, ok := value.(string); ok {
			args = append(args, prefix+key, strValue)
		}
	}
	return args
}
