package controllers

import (
	"errors"
	"net/http"
	"os"

	"lite-nas/shared/fileio"
	sharedlogger "lite-nas/shared/logger"
)

// StaticFiles groups the packaged browser resources served by the gateway.
type StaticFiles struct {
	IndexHTML fileio.Reader
	IndexCSS  fileio.Reader
	IndexJS   fileio.Reader
	Favicon   fileio.Reader
}

// StaticController serves the packaged browser assets owned by the gateway.
type StaticController struct {
	files  StaticFiles
	logger sharedlogger.Logger
}

// NewStaticController creates a StaticController.
//
// Parameters:
//   - files: packaged frontend file readers used by the handlers
//   - logger: application logger used when a packaged resource cannot be read
func NewStaticController(files StaticFiles, logger sharedlogger.Logger) StaticController {
	return StaticController{
		files:  files,
		logger: logger,
	}
}

// ServeIndex serves the packaged HTML entrypoint for the browser UI.
func (c StaticController) ServeIndex(writer http.ResponseWriter, request *http.Request) {
	c.serveFile(writer, c.files.IndexHTML, "index resource", "text/html; charset=utf-8")
}

// ServeIndexCSS serves the packaged stylesheet for the browser UI.
func (c StaticController) ServeIndexCSS(writer http.ResponseWriter, request *http.Request) {
	c.serveFile(writer, c.files.IndexCSS, "index stylesheet", "text/css; charset=utf-8")
}

// ServeIndexJS serves the packaged JavaScript bundle for the browser UI.
func (c StaticController) ServeIndexJS(writer http.ResponseWriter, request *http.Request) {
	c.serveFile(writer, c.files.IndexJS, "index script", "application/javascript; charset=utf-8")
}

// ServeFavicon serves the packaged favicon for the browser UI.
func (c StaticController) ServeFavicon(writer http.ResponseWriter, request *http.Request) {
	c.serveFile(writer, c.files.Favicon, "favicon", "image/x-icon")
}

func (c StaticController) serveFile(
	writer http.ResponseWriter,
	reader fileio.Reader,
	resourceName string,
	contentType string,
) {
	data, err := reader.Read()
	if err != nil {
		c.logger.Error("failed to load static resource", "resource", resourceName, "error", err.Error())

		if errors.Is(err, os.ErrNotExist) {
			http.Error(writer, resourceName+" not found", http.StatusNotFound)
			return
		}

		http.Error(writer, "failed to load "+resourceName, http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", contentType)
	_, _ = writer.Write(data)
}
