package main

import (
	"context"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

/*
# Takehome Notes

https://www.dicomlibrary.com/dicom/dicom-tags/
https://dicomiseasy.blogspot.com/2011/10/introduction-to-dicom-chapter-1.html
https://www.digitalocean.com/community/tutorials/how-to-make-an-http-server-in-go#prerequisites
https://github.com/suyashkumar/dicom

// TODO: mis-read requirements as all one endpoint. split out each functionality to separate endpoints DUH

- POST /dicom
  - extract header attributes
  - convert to PNG
  - return location header with resource URL https://stackoverflow.com/questions/1829875/is-it-ok-by-rest-to-return-content-after-post
  - if client include query param (DICOM Tag), return the header attribute as JSON
- GET /dicom/:id
	- use query params to determine type of return? imageType='png', default original dicom?
*/

func main() {
	r := chi.NewRouter()
	r.Get("/", rootHandler)
	r.Route("/dicom-files", func(r chi.Router) {
		r.Post("/", createDicomResource)
		r.Route("/{dicomFileID}", func(r chi.Router) {
			r.Use(dicomFileCtx)
			r.Get("/", getDicomResource)
			// r.Get("/file", getDicomFile)
			// r.Get("/image", getDicomImage)
		})
	})

	http.ListenAndServe(":3001", r)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("welcome"))
}

func createDicomResource(w http.ResponseWriter, r *http.Request) {
	id := uuid.New()

	path := filepath.Join("images", id.String())
	file, _ := os.Create(path)
	uploadedFile, _, _ := r.FormFile("file")
	io.Copy(file, uploadedFile)

	params := r.URL.Query().Get("tag")
	tag := parseDicomTagParams(params)
	dataset, _ := dicom.ParseFile(filepath.Join("images", id.String()), nil)
	element, _ := dataset.FindElementByTag(tag)

	convertDicomToPng(id.String())

	// TODO: return value
	w.Write([]byte("TODO create file"))
}

func parseDicomTagParams(params string) tag.Tag {
	params = strings.TrimPrefix(params, "(")
	params = strings.TrimSuffix(params, ")")
	tagParts := strings.Split(params, ",")

	group, _ := strconv.ParseUint(tagParts[0], 10, 16)
	element, _ := strconv.ParseUint(tagParts[1], 10, 16)

	return tag.Tag{Group: uint16(group), Element: uint16(element)}
}

// TODO: prevent reading from disk
// tried but failed to parse the form upload file and the tempfile in createDicomFile
func convertDicomToPng(id string) {
	dataset, _ := dicom.ParseFile(filepath.Join("images", id), nil)
	pixelDataElement, _ := dataset.FindElementByTag(tag.PixelData)
	pixelDataInfo := dicom.MustGetPixelDataInfo(pixelDataElement.Value)

	for i, fr := range pixelDataInfo.Frames {
		img, _ := fr.GetImage()
		f, _ := os.Create(filepath.Join("images", fmt.Sprintf("%s_%d.png", id, i)))
		_ = png.Encode(f, img)
		_ = f.Close()
	}
}

func dicomFileCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dicomFileID := chi.URLParam(r, "dicomFileID")
		// TODO: more details
		dicomFile := dicomFileID
		ctx := context.WithValue(r.Context(), "dicomFile", dicomFile)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getDicomResource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	value, _ := ctx.Value("dicomFile").(string)

	// TODO: implement query param here

	// TODO: return JSON dicom dataset
	fmt.Println("get handler", value)
	w.Write([]byte("TODO get file"))
}
