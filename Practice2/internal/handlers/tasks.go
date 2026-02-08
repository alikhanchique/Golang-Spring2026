package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"kbtu.practice2.alikh/internal/storage"
)

type TaskHandler struct {
	Store *storage.TaskStore
}

func NewTaskHandler(store *storage.TaskStore) *TaskHandler {
	return &TaskHandler{Store: store}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *TaskHandler) Tasks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPost:
		h.handlePost(w, r)
	case http.MethodPatch:
		h.handlePatch(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
	}
}

func (h *TaskHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")

	if idStr == "" {
		tasks := h.Store.GetAll()
		writeJSON(w, http.StatusOK, tasks)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid id",
		})
		return
	}

	task, err := h.Store.GetByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "task not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
	}

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid title",
		})
		return
	}

	title := strings.TrimSpace(body.Title)
	if title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid title",
		})
		return
	}

	task := h.Store.Create(title)
	writeJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) handlePatch(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid id",
		})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid id",
		})
		return
	}

	var body struct {
		Done *bool `json:"done"`
	}

	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil || body.Done == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "done must be boolean",
		})
		return
	}

	err = h.Store.UpdateDone(id, *body.Done)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "task not found",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{
		"updated": true,
	})
}
