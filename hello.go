package main

import (
	"context"
	"fmt"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"

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
		r.Post("/", createDicomFile)
		r.Route("/{dicomFileID}", func(r chi.Router) {
			r.Use(dicomFileCtx)
			r.Get("/", getDicomFile)
		})
	})

	http.ListenAndServe(":3001", r)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("welcome"))
}

func createDicomFile(w http.ResponseWriter, r *http.Request) {
	id := uuid.New()

	path := filepath.Join("images", id.String())
	file, _ := os.Create(path)
	uploadedFile, _, _ := r.FormFile("file")
	io.Copy(file, uploadedFile)

	convertDicomToPng(path)

	// TODO: parse query param
	// TODO: return value
	w.Write([]byte("TODO create file"))
}

// TODO: prevent reading from disk
// tried but failed to parse the form upload file and the tempfile in createDicomFile
func convertDicomToPng(path string) {
	dataset, _ := dicom.ParseFile(path, nil)
	pixelDataElement, _ := dataset.FindElementByTag(tag.PixelData)
	pixelDataInfo := dicom.MustGetPixelDataInfo(pixelDataElement.Value)
	fmt.Println(pixelDataInfo)
	for i, fr := range pixelDataInfo.Frames {
		img, _ := fr.GetImage()
		// TODO: same name, same directory as dicom
		f, _ := os.Create(fmt.Sprintf("image_%d.jpg", i))
		// TODO: convert to png
		_ = jpeg.Encode(f, img, &jpeg.Options{Quality: 100})
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

func getDicomFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	value, _ := ctx.Value("dicomFile").(string)

	// TODO: what to return??
	fmt.Println("get handler", value)
	w.Write([]byte("TODO get file"))
}
