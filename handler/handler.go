package handler

import (
	"1/checker"
	"1/pdf"
	"1/storage"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type handler struct {
	store storage.Storage
	check *checker.LinkChecker
}

func NewHandler(store storage.Storage, checker *checker.LinkChecker) *handler {
	if checker == nil {
		log.Fatalf("LinkChecker is nil")
	}
	return &handler{
		store: store,
		check: checker,
	}
}

func NewRouter(h *handler) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/api/check", h.CheckLinks).Methods("POST")
	r.HandleFunc("/api/report", h.GetReport).Methods("POST")
	r.HandleFunc("/api/status/{id}", h.GetStatus).Methods("GET")
	r.HandleFunc("/health", h.HealthCheck).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Content-Type", "X-Requested-With"},
	})

	return c.Handler(r)
}

type CheckResponse struct {
	Links    map[string]string `json:"links"`
	LinksNum int               `json:"links_num"`
}
type CheckRequest struct {
	Links []string `json:"links"`
}
type ReportRequest struct {
	LinksList []int `json:"links_list"`
}

func (h *handler) CheckLinks(w http.ResponseWriter, r *http.Request) {
	var req CheckRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(req.Links) == 0 {
		http.Error(w, "Links is empty", http.StatusBadRequest)
		return
	}

	setID, err := h.store.SaveLinks(req.Links)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.check.CheckLinksAsync(setID, req.Links)

	response := CheckResponse{
		Links:    make(map[string]string),
		LinksNum: setID,
	}

	for _, link := range req.Links {
		response.Links[link] = "pending"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	set, ok := h.store.GetLinkSet(id)
	if !ok {
		http.Error(w, fmt.Sprintf("LinkSet %d not found", id), http.StatusNotFound)
		return
	}

	response := CheckResponse{
		Links:    make(map[string]string),
		LinksNum: id,
	}
	for _, link := range set.Links {
		response.Links[link.URL] = string(link.Status)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *handler) GetReport(w http.ResponseWriter, r *http.Request) {
	var req ReportRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(req.LinksList) == 0 {
		http.Error(w, "LinksList is empty", http.StatusBadRequest)
		return
	}

	sets, err := h.store.GetLinkSets(req.LinksList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pdfBytes, err := pdf.GenerateReport(sets)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=report_%d.pdf\"", time.Now().Unix()))
	w.Write(pdfBytes)
}

func (h *handler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "OK",
		"time":   time.Now().Format(time.RFC3339),
	})
}
