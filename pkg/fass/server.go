package fass

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func courseExerciseFromRequest(r *http.Request) (course Course, exercise Exercise, err error) {
	vars := mux.Vars(r)

	course, err = LoadCourse(vars["course"])
	if err != nil {
		err = errors.New("course not found")
		return
	}

	exercise, found := course.Exercises[vars["exercise"]]
	if !found {
		err = errors.New("exercise not found")
		return
	}

	return
}

func handleBuildStatus(w http.ResponseWriter, r *http.Request) {
	_, exercise, err := courseExerciseFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// TODO check token
	submissionFilename := "t0k3n.zip"

	buildOutput, err := exercise.GetBuildOutput(submissionFilename)
	if err != nil {
		http.Error(w, "no build output available", http.StatusNotFound)
		return
	}

	io.Copy(w, buildOutput)

	if exercise.WasBuildSuccessful(submissionFilename) {
		fmt.Fprintf(w, "\nBuild Succeeded!\n")
	} else {
		fmt.Fprintf(w, "\nBuild Failed!\n")
	}
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	course, exercise, err := courseExerciseFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// TODO check token

	submission, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !isZIP(submission) {
		http.Error(w, "w=Wrong content type, application/zip required.", http.StatusBadRequest)
		return
	}

	submissionFilename := "t0k3n.zip"

	err = exercise.StoreSubmission(submission, submissionFilename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Invoking build:", course.Identifier, exercise.Identifier, submissionFilename)
	go invokeBuild(exercise, submissionFilename)

	fmt.Fprintln(w, "Upload successful")
}

func invokeBuild(exercise Exercise, submissionFilename string) {
	err := exercise.BuildSubmission(submissionFilename)
	if err != nil {
		log.Println(err)
	}
}

func isZIP(file io.Reader) bool {
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return false
	}
	return http.DetectContentType(buffer) == "application/zip"
}

func Serve(addr string) {
	router := mux.NewRouter()
	router.HandleFunc("/{course}/{exercise}", handleBuildStatus).Methods("GET")
	router.HandleFunc("/{course}/{exercise}", handleUpload).Methods("POST")

	http.Handle("/", router)

	log.Printf("Listening on %s\n", addr)
	http.ListenAndServe(addr, nil)
}
