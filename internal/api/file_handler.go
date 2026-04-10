package api

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sakkshm/bastion/internal/session"
)

func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	sess, ok := r.Context().Value(SessionContextKey).(*session.Session)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "session context missing")
		return
	}

	// Update session data
	h.Engine.Sessions.Touch(sess.ID)

	// limit file upload size (in MBs)
	maxSizeMB := h.Engine.Config.FileSystem.MaxUploadSize
	err := r.ParseMultipartForm(int64(maxSizeMB) << 20)
	if err != nil {
		h.Engine.Logger.Error(
			"Cannot parse form",
			"session_id", sess.ID,
			"error", err,
		)
		writeJSONError(w, http.StatusBadRequest, "Cannot parse form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.Engine.Logger.Error(
			"Failed to retrieve file",
			"session_id", sess.ID,
			"error", err,
		)
		writeJSONError(w, http.StatusBadRequest, "Failed to retrieve file")
		return
	}
	defer file.Close()

	// validate file size
	if header.Size > int64(maxSizeMB)<<20 {
		writeJSONError(w, http.StatusBadRequest, "File too large")
		return
	}

	// Get metadata
	metadataStr := r.FormValue("metadata")
	if metadataStr == "" {
		writeJSONError(w, http.StatusBadRequest, "Metadata required")
		return
	}

	var meta UploadMetadata
	err = json.Unmarshal([]byte(metadataStr), &meta)
	if err != nil {
		h.Engine.Logger.Error(
			"Invalid metadata",
			"session_id", sess.ID,
			"error", err,
		)
		writeJSONError(w, http.StatusBadRequest, "Invalid metadata")
		return
	}

	// check if safe
	meta.Path = filepath.Clean(meta.Path)
	uploadPath, err := sess.FileSystem.SafePath(meta.Path)
	if err != nil {
		h.Engine.Logger.Error(
			"Invalid upload path",
			"session_id", sess.ID,
			"path", meta.Path,
			"error", err,
		)
		writeJSONError(w, http.StatusBadRequest, "Invalid upload path")
		return
	}

	// prevent overwrite
	if _, err := os.Stat(uploadPath); err == nil {
		writeJSONError(w, http.StatusConflict, "File already exists")
		return
	}

	// ensure directory exists
	err = os.MkdirAll(filepath.Dir(uploadPath), 0755)
	if err != nil {
		h.Engine.Logger.Error(
			"Cannot create directory",
			"session_id", sess.ID,
			"path", uploadPath,
			"error", err,
		)
		writeJSONError(w, http.StatusInternalServerError, "Cannot create directory")
		return
	}

	// create upload dest
	dst, err := os.OpenFile(uploadPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0644)
	if err != nil {
		h.Engine.Logger.Error(
			"Unable to create upload path",
			"session_id", sess.ID,
			"path", uploadPath,
			"error", err,
		)
		writeJSONError(w, http.StatusInternalServerError, "Unable to create upload path")
		return
	}
	defer dst.Close()

	// copy content
	_, err = io.Copy(dst, file)
	if err != nil {
		dst.Close()
		_ = os.Remove(uploadPath) // cleanup partial file

		h.Engine.Logger.Error(
			"Error writing file",
			"session_id", sess.ID,
			"path", uploadPath,
			"error", err,
		)
		writeJSONError(w, http.StatusInternalServerError, "Error writing file")
		return
	}

	// success response
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(FileUploadResponse{
		Status: "success",
		Path:   meta.Path,
	})
}

func (h *Handler) DownloadHandler(w http.ResponseWriter, r *http.Request) {

	sess, ok := r.Context().Value(SessionContextKey).(*session.Session)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "session context missing")
		return
	}

	// Update session data
	h.Engine.Sessions.Touch(sess.ID)

	// extract query param into struct
	var req DownloadRequest
	req.Path = r.URL.Query().Get("path")

	if req.Path == "" {
		writeJSONError(w, http.StatusBadRequest, "path required")
		return
	}

	// check if safe
	req.Path = filepath.Clean(req.Path)
	downloadPath, err := sess.FileSystem.SafePath(req.Path)
	if err != nil {
		h.Engine.Logger.Error(
			"Invalid download path",
			"session_id", sess.ID,
			"path", req.Path,
			"error", err,
		)
		writeJSONError(w, http.StatusBadRequest, "Invalid download path")
		return
	}

	// check if file exists
	info, err := os.Stat(downloadPath)
	if err != nil {
		if os.IsNotExist(err) {
			h.Engine.Logger.Error(
				"File does not exist",
				"session_id", sess.ID,
				"path", req.Path,
			)
			writeJSONError(w, http.StatusNotFound, "File does not exist")
		} else {
			h.Engine.Logger.Error(
				"Failed to access file",
				"session_id", sess.ID,
				"path", req.Path,
				"error", err,
			)
			writeJSONError(w, http.StatusInternalServerError, "Failed to access file")
		}
		return
	}

	if info.IsDir() {
		writeJSONError(w, http.StatusBadRequest, "Path is a directory")
		return
	}

	// force download
	filename := filepath.Base(downloadPath)
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)

	// set mime type
	ext := filepath.Ext(downloadPath)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", mimeType)

	http.ServeFile(w, r, downloadPath)
}

func (h *Handler) DeleteHandler(w http.ResponseWriter, r *http.Request) {

	sess, ok := r.Context().Value(SessionContextKey).(*session.Session)
	if !ok {
		writeJSONError(w, http.StatusInternalServerError, "session context missing")
		return
	}

	// Update session data
	h.Engine.Sessions.Touch(sess.ID)

	// extract query param into struct
	var req DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer r.Body.Close()

	if req.Path == "" {
		writeJSONError(w, http.StatusBadRequest, "path required")
		return
	}

	// check if safe
	req.Path = filepath.Clean(req.Path)
	deletePath, err := sess.FileSystem.SafePath(req.Path)
	if err != nil {
		h.Engine.Logger.Error(
			"Invalid delete path",
			"session_id", sess.ID,
			"path", req.Path,
			"error", err,
		)
		writeJSONError(w, http.StatusBadRequest, "Invalid delete path")
		return
	}

	if deletePath == sess.FileSystem.Mount {
		writeJSONError(w, http.StatusBadRequest, "cannot delete root directory")
		return
	}

	// check if file exists
	info, err := os.Stat(deletePath)
	if err != nil {
		if os.IsNotExist(err) {
			h.Engine.Logger.Error(
				"File does not exist",
				"session_id", sess.ID,
				"path", req.Path,
			)
			writeJSONError(w, http.StatusNotFound, "File does not exist")
		} else {
			h.Engine.Logger.Error(
				"Failed to access file",
				"session_id", sess.ID,
				"path", req.Path,
				"error", err,
			)
			writeJSONError(w, http.StatusInternalServerError, "Failed to access file")
		}
		return
	}

	if info.IsDir() {
		err = os.RemoveAll(deletePath)
	} else {
		err = os.Remove(deletePath)
	}

	if err != nil {
		h.Engine.Logger.Error(
			"Failed to delete file",
			"session_id", sess.ID,
			"path", req.Path,
			"error", err,
		)
		writeJSONError(w, http.StatusInternalServerError, "Failed to delete file")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(FileDeleteResponse{
		Status: "deleted successfully",
	})
}
