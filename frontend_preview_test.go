//go:build ui_preview

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// TestFrontendPreview is an opt-in localhost bridge for browser inspection of
// real M6-006 records. It is excluded from the application and ordinary tests.
func TestFrontendPreview(t *testing.T) {
	root := os.Getenv("ATROPATEN_UI_DEMO_ROOT")
	if root == "" { t.Skip("set ATROPATEN_UI_DEMO_ROOT to a marked disposable demo root") }
	root, err := filepath.Abs(root)
	if err != nil { t.Fatal(err) }
	if !strings.HasPrefix(root, os.TempDir()+string(os.PathSeparator)) { t.Fatal("preview requires a disposable temporary root") }
	if _, err := os.Stat(filepath.Join(root, ".atropaten-demo.json")); err != nil { t.Fatal("preview requires an M6-006 marker", err) }
	t.Setenv("ATROPATEN_DATA_DIR", root)
	app := NewApp()
	app.startup(context.Background())
	if app.startupError != nil { t.Fatal(app.startupError) }
	defer app.shutdown(context.Background())
	http.HandleFunc("/call/", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Origin") != "http://127.0.0.1:5178" { http.Error(w, "preview origin required", 403); return }
		w.Header().Set("Access-Control-Allow-Origin", "http://127.0.0.1:5178")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" { return }
		if r.Method != "POST" { http.Error(w, "POST required", 405); return }
		name := strings.TrimPrefix(r.URL.Path, "/call/")
		// No backup/restore, filesystem dialogs, or native runtime operations.
		if strings.Contains(name, "Backup") || strings.Contains(name, "File") && name != "ListAttachments" { http.Error(w, "native operation unavailable in browser preview", 400); return }
		method := reflect.ValueOf(app).MethodByName(name)
		if !method.IsValid() { http.Error(w, "unknown bridge method", 404); return }
		var raw []json.RawMessage
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&raw); err != nil || len(raw) != method.Type().NumIn() { http.Error(w, "invalid arguments", 400); return }
		args := make([]reflect.Value, len(raw))
		for i := range raw { value := reflect.New(method.Type().In(i)); if err := json.Unmarshal(raw[i], value.Interface()); err != nil { http.Error(w, err.Error(), 400); return }; args[i] = value.Elem() }
		result := method.Call(args)
		var data any
		for _, value := range result { if value.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) { if !value.IsNil() { http.Error(w, value.Interface().(error).Error(), 400); return } } else { data = value.Interface() } }
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil { t.Log(err) }
	})
	t.Log("M6-006 preview listening at 127.0.0.1:4174")
	if err := http.ListenAndServe("127.0.0.1:4174", nil); err != nil { t.Fatal(err) }
}
