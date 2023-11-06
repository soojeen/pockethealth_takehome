package main

import (
	"context"
	"encoding/json"
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

type DicomFile struct {
	ID      string        `json:"id"`
	Dataset dicom.Dataset `json:"dataset"`
}

/*
# Takehome Notes

https://www.dicomlibrary.com/dicom/dicom-tags/
https://dicomiseasy.blogspot.com/2011/10/introduction-to-dicom-chapter-1.html
https://www.digitalocean.com/community/tutorials/how-to-make-an-http-server-in-go#prerequisites
https://github.com/suyashkumar/dicom

// TODO: mis-read requirements as all one endpoint. split out each functionality to separate endpoints DUH

- POST /dicom
  - convert to PNG
  - if client include query param (DICOM Tag), return the header attribute as JSON
- GET /dicom/:id
*/

func main() {
	r := chi.NewRouter()

	r.Get("/", rootHandler)
	r.Route("/dicom-files", func(r chi.Router) {
		r.Post("/", createDicomResource)
		r.Route("/{dicomFileID}", func(r chi.Router) {
			r.Use(dicomFileCtx)
			r.Get("/", getDicomResource)
			r.Get("/file", getDicomFile)
			r.Get("/image", getDicomImage)
		})
	})

	http.ListenAndServe(":3001", r)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("welcome to the Dicom file server"))
}

func createDicomResource(w http.ResponseWriter, r *http.Request) {
	id := uuid.New()

	path := filepath.Join("images", id.String())
	file, _ := os.Create(path)
	uploadedFile, _, _ := r.FormFile("file")
	io.Copy(file, uploadedFile)

	// NOTE: considered returning full dataset as resource, but it is very large.
	// client can get any further data based on the Location header url
	w.Header().Set("Location", filepath.Join("dicom-files", id.String()))
	w.WriteHeader(http.StatusCreated)
}

func dicomFileCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dicomFileID := chi.URLParam(r, "dicomFileID")
		dataset, _ := dicom.ParseFile(filepath.Join("images", dicomFileID), nil)
		ctx := context.WithValue(r.Context(), "dicomId", dicomFileID)
		ctx = context.WithValue(r.Context(), "dicomDataset", dataset)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func parseDicomTagParams(params string) tag.Tag {
	params = strings.TrimPrefix(params, "(")
	params = strings.TrimSuffix(params, ")")
	tagParts := strings.Split(params, ",")

	group, _ := strconv.ParseUint(tagParts[0], 10, 16)
	element, _ := strconv.ParseUint(tagParts[1], 10, 16)

	return tag.Tag{Group: uint16(group), Element: uint16(element)}
}

func getDicomResource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dataset, _ := ctx.Value("dicomDataset").(dicom.Dataset)

	params := r.URL.Query().Get("tag")

	if params != "" {
		tag := parseDicomTagParams(params)
		element, _ := dataset.FindElementByTag(tag)
		jData, _ := json.Marshal(element)
		w.Write(jData)
	} else {
		jData, _ := json.Marshal(dataset)
		w.Write(jData)
	}
}

func getDicomFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := ctx.Value("dicomId").(string)
	path := filepath.Join("images", id)

	// TODO: return dicom file
	fmt.Println(path)
}

func convertDicomToPng(dataset dicom.Dataset, id string) []string {
	pixelDataElement, _ := dataset.FindElementByTag(tag.PixelData)
	pixelDataInfo := dicom.MustGetPixelDataInfo(pixelDataElement.Value)

	filePaths := make([]string, len(pixelDataInfo.Frames))
	for i, fr := range pixelDataInfo.Frames {
		img, _ := fr.GetImage()
		path := filepath.Join("images", fmt.Sprintf("%s_%d.png", id, i))
		f, _ := os.Create(path)
		_ = png.Encode(f, img)
		_ = f.Close()
		filePaths[i] = path
	}

	return filePaths
}

func getDicomImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := ctx.Value("dicomId").(string)
	dataset, _ := ctx.Value("dicomDataset").(dicom.Dataset)

	filePaths := convertDicomToPng(dataset, id)

	if len(filePaths) == 1 {
		// TODO: return image file
	} else {
		// TODO: dicom library suggests each dicom file could contain a range of image frames
		// but in practice, only seeing one image. implement this when we need to return zip of multiple files
	}
}
