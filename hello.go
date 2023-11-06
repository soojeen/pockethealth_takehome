package main

import (
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
*/

func main() {
	r := chi.NewRouter()
	r.Get("/", rootHandler)

	// fmt.Println("Hello, World!")

	http.ListenAndServe(":3001", r)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("welcome"))
}
