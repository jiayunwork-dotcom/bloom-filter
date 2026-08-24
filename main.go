package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"bloom-filter/internal/codec"
	"bloom-filter/internal/filter"
	"bloom-filter/internal/store"
)

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "bloom-filter: "+format+"\n", args...)
	os.Exit(1)
}

// reorderFlags moves all "-flag [value]" pairs to the front so that a positional
// argument may safely appear before flags.
func reorderFlags(args []string) []string {
	var flags, pos []string
	i := 0
	for i < len(args) {
		a := args[i]
		if strings.HasPrefix(a, "-") {
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") && !strings.Contains(a, "=") {
				flags = append(flags, a, args[i+1])
				i += 2
			} else {
				flags = append(flags, a)
				i++
			}
		} else {
			pos = append(pos, a)
			i++
		}
	}
	return append(flags, pos...)
}

// stringSlice is a repeatable flag that splits each value by comma/space/tab.
type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ",") }

func (s *stringSlice) Set(v string) error {
	f := func(r rune) bool { return r == ',' || r == ' ' || r == '\t' }
	for _, p := range strings.FieldsFunc(v, f) {
		if p != "" {
			*s = append(*s, p)
		}
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fatal("usage: bloom-filter <serve|add|check|info|store-add|store-check|checkpoint> ...")
	}
	cmd := os.Args[1]
	args := reorderFlags(os.Args[2:])
	switch cmd {
	case "serve":
		runServe(args)
	case "add":
		runAdd(args)
	case "check":
		runCheck(args)
	case "info":
		runInfo(args)
	case "store-add":
		runStoreAdd(args)
	case "store-check":
		runStoreCheck(args)
	case "checkpoint":
		runCheckpoint(args)
	default:
		fatal("unknown command %q (want serve|add|check|info|store-add|store-check|checkpoint)", cmd)
	}
}

const (
	defaultM = 1 << 16 // 65536 bits
	defaultK = 7
)

func runAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	file := fs.String("f", "", "filter file (v1 or v2 format)")
	var items stringSlice
	fs.Var(&items, "items", "items to add (comma/space separated, repeatable)")
	if err := fs.Parse(args); err != nil {
		fatal("%v", err)
	}
	if *file == "" {
		fatal("add requires -f <file>")
	}
	items = append(items, fs.Args()...)
	if len(items) == 0 {
		fatal("add requires at least one item (-items or positional)")
	}

	var f *filter.BloomFilter
	if data, err := os.ReadFile(*file); err == nil {
		f, err = codec.Unmarshal(data)
		if err != nil {
			fatal("unmarshal existing filter: %v", err)
		}
	} else if os.IsNotExist(err) {
		f, err = filter.New(defaultM, defaultK)
		if err != nil {
			fatal("%v", err)
		}
	} else {
		fatal("read %s: %v", *file, err)
	}

	for _, it := range items {
		f.Add([]byte(it))
	}
	if err := os.WriteFile(*file, codec.Marshal(f), 0644); err != nil {
		fatal("write %s: %v", *file, err)
	}
	fmt.Printf("added %d item(s) to %s\n", len(items), *file)
}

func runCheck(args []string) {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	file := fs.String("f", "", "filter file")
	item := fs.String("item", "", "item to check")
	if err := fs.Parse(args); err != nil {
		fatal("%v", err)
	}
	if *file == "" {
		fatal("check requires -f <file>")
	}
	if *item == "" {
		fatal("check requires -item <x>")
	}
	data, err := os.ReadFile(*file)
	if err != nil {
		fatal("read %s: %v", *file, err)
	}
	f, err := codec.Unmarshal(data)
	if err != nil {
		fatal("unmarshal: %v", err)
	}
	if f.Test([]byte(*item)) {
		fmt.Println("present")
	} else {
		fmt.Println("absent")
	}
}

func runInfo(args []string) {
	fs := flag.NewFlagSet("info", flag.ContinueOnError)
	file := fs.String("f", "", "filter file")
	if err := fs.Parse(args); err != nil {
		fatal("%v", err)
	}
	if *file == "" {
		fatal("info requires -f <file>")
	}
	data, err := os.ReadFile(*file)
	if err != nil {
		fatal("read %s: %v", *file, err)
	}
	ver, err := codec.Version(data)
	if err != nil {
		fatal("version: %v", err)
	}
	f, err := codec.Unmarshal(data)
	if err != nil {
		fatal("unmarshal: %v", err)
	}
	fmt.Printf("version=%d m=%d k=%d bytes=%d\n", ver, f.M(), f.K(), len(data))
	fmt.Printf("max_insertions(fpr=0.01)=%d\n", f.MaxInsertions(0.01))
}

func runStoreAdd(args []string) {
	fs := flag.NewFlagSet("store-add", flag.ContinueOnError)
	file := fs.String("f", "", "store file (.blm)")
	m := fs.Uint("m", uint(defaultM), "filter bits (only for new store)")
	k := fs.Uint("k", uint(defaultK), "hash functions (only for new store)")
	var items stringSlice
	fs.Var(&items, "items", "items to add (comma/space separated, repeatable)")
	if err := fs.Parse(args); err != nil {
		fatal("%v", err)
	}
	if *file == "" {
		fatal("store-add requires -f <file>")
	}
	items = append(items, fs.Args()...)
	if len(items) == 0 {
		fatal("store-add requires at least one item")
	}

	var s *store.Store
	var err error
	if _, serr := os.Stat(*file); os.IsNotExist(serr) {
		s, err = store.Create(*file, *m, *k)
	} else {
		s, err = store.Open(*file)
	}
	if err != nil {
		fatal("open store: %v", err)
	}
	defer s.Close()

	for _, it := range items {
		if err := s.Add([]byte(it)); err != nil {
			fatal("store add: %v", err)
		}
	}
	fmt.Printf("store-added %d item(s) to %s\n", len(items), *file)
}

func runStoreCheck(args []string) {
	fs := flag.NewFlagSet("store-check", flag.ContinueOnError)
	file := fs.String("f", "", "store file (.blm)")
	item := fs.String("item", "", "item to check")
	if err := fs.Parse(args); err != nil {
		fatal("%v", err)
	}
	if *file == "" {
		fatal("store-check requires -f <file>")
	}
	if *item == "" {
		fatal("store-check requires -item <x>")
	}

	s, err := store.Open(*file)
	if err != nil {
		fatal("open store: %v", err)
	}
	defer s.Close()

	if s.Test([]byte(*item)) {
		fmt.Println("present")
	} else {
		fmt.Println("absent")
	}
}

func runCheckpoint(args []string) {
	fs := flag.NewFlagSet("checkpoint", flag.ContinueOnError)
	file := fs.String("f", "", "store file (.blm)")
	if err := fs.Parse(args); err != nil {
		fatal("%v", err)
	}
	if *file == "" {
		fatal("checkpoint requires -f <file>")
	}
	s, err := store.Open(*file)
	if err != nil {
		fatal("open store: %v", err)
	}
	defer s.Close()
	if err := s.Checkpoint(); err != nil {
		fatal("checkpoint: %v", err)
	}
	fmt.Printf("checkpoint written to %s\n", *file)
}


func runServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	addr := fs.String("addr", ":8080", "listen address")
	file := fs.String("f", "data.blm", "store file (.blm)")
	m := fs.Uint("m", uint(defaultM), "filter bits (only for new store)")
	k := fs.Uint("k", uint(defaultK), "hash functions (only for new store)")
	if err := fs.Parse(args); err != nil {
		fatal("%v", err)
	}

	var s *store.Store
	var err error
	if _, serr := os.Stat(*file); os.IsNotExist(serr) {
		s, err = store.Create(*file, *m, *k)
	} else {
		s, err = store.Open(*file)
	}
	if err != nil {
		fatal("open store: %v", err)
	}

	srv := &server{store: s, file: *file}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/add", srv.handleAdd)
	mux.HandleFunc("/api/check", srv.handleCheck)
	mux.HandleFunc("/api/checkpoint", srv.handleCheckpoint)
	mux.HandleFunc("/api/info", srv.handleInfo)
	mux.HandleFunc("/health", srv.handleHealth)

	fmt.Printf("bloom-filter serving on %s (store=%s)\n", *addr, *file)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		fatal("listen: %v", err)
	}
}

type server struct {
	store *store.Store
	file  string
	mu    sync.RWMutex
}

func (sv *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (sv *server) handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Items []string `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(req.Items) == 0 {
		http.Error(w, "items required", http.StatusBadRequest)
		return
	}
	sv.mu.Lock()
	defer sv.mu.Unlock()
	for _, it := range req.Items {
		if err := sv.store.Add([]byte(it)); err != nil {
			http.Error(w, "store add: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"added": len(req.Items)})
}

func (sv *server) handleCheck(w http.ResponseWriter, r *http.Request) {
	item := r.URL.Query().Get("item")
	if item == "" {
		http.Error(w, "item query param required", http.StatusBadRequest)
		return
	}
	sv.mu.RLock()
	present := sv.store.Test([]byte(item))
	sv.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"present": present})
}

func (sv *server) handleCheckpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sv.mu.Lock()
	defer sv.mu.Unlock()
	if err := sv.store.Checkpoint(); err != nil {
		http.Error(w, "checkpoint: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "checkpoint written"})
}

func (sv *server) handleInfo(w http.ResponseWriter, r *http.Request) {
	sv.mu.RLock()
	f := sv.store.Filter()
	sv.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"m":              f.M(),
		"k":              f.K(),
		"max_insert_001": f.MaxInsertions(0.01),
	})
}
