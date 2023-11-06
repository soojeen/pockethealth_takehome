package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
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

	// fmt.Println("Hello, World!")

	http.ListenAndServe(":3001", r)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("welcome"))
}

func createDicomFile(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("TODO create file"))
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
	fmt.Println("Hello, World!", value)
	w.Write([]byte("TODO get file"))
}
