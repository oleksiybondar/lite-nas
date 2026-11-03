package main

import (
	"net/http"
	"os"

	"lite-nas/shared/fileaccess"
)

// statResource returns the kind of one source path after File Service authorization.
func (s *fileService) statResource(writer http.ResponseWriter, request *http.Request, subject fileaccess.Subject) {
	if request.Method != http.MethodGet {
		writeError(writer, http.StatusMethodNotAllowed, "method_not_allowed")
		return
	}
	if request.URL.RawQuery != "" {
		writeError(writer, http.StatusBadRequest, "invalid_query")
		return
	}
	resolved, scope, ok := s.resolveReadScope(writer, request, subject, "/stat")
	if !ok {
		return
	}
	info, err := os.Stat(resolved)
	if err != nil {
		writeError(writer, http.StatusNotFound, "not_found")
		return
	}
	kind := ""
	if info.IsDir() && scope.AllowsDirectory(resolved) {
		kind = "directory"
	} else if info.Mode().IsRegular() && scope.AllowsFile(resolved) {
		kind = "file"
	}
	if kind == "" {
		decisionError(writer, fileaccess.ErrDenied)
		return
	}
	if err := s.auth.Read(request.Context(), subject, resolved); err != nil {
		decisionError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, map[string]string{"path": resolved, "kind": kind})
}
