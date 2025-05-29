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
	logger := Logger.With("method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr)
	if r.URL.Path == "/_health" {
		w.Write([]byte("OK"))
		return
	}
	defer r.Body.Close()
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

	p, ok := body["package"].(map[string]interface{})
	if !ok {
		logger.Error("Package field is missing or not an object")
		http.Error(w, "Package field is missing or not an object", http.StatusBadRequest)
		return
	}
	args := toArgs(p, "")
	reg, ok := p["registry"].(map[string]interface{})
	if !ok {
		logger.Error("Registry field is missing or not an object")
		http.Error(w, "Registry field is missing or not an object", http.StatusBadRequest)
		return
	}
	args = append(args, toArgs(reg, "registry.")...)
	v, ok := p["package_version"].(map[string]interface{})
	if !ok {
		logger.Error("Package version field is missing or not an object")
		http.Error(w, "Package version field is missing or not an object", http.StatusBadRequest)
		return
	}
	args = append(args, toArgs(v, "package_version.")...)
	args = append(args, "tag", v["container_metadata"].(map[string]interface{})["tag"].(map[string]interface{})["name"].(string))

	argsAny := make([]any, len(args))
	for i, v := range args {
		argsAny[i] = v
	}
	logger.Info("package created", argsAny...)
	w.Write([]byte("Infrastructure endpoint"))
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
