package fass

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func tokenFromRequest(r *http.Request) Token {
	return r.Header.Get("X-Auth-Token")
}

func submissionFilenameFromRequest(r *http.Request) string {
	return tokenFromRequest(r) + ".zip"
}

func courseFromRequest(r *http.Request) (course Course, err error) {
	vars := mux.Vars(r)

	course, err = LoadCourse(vars["course"])
	if err != nil {
		err = errors.New("course not found")
	}

	return
}

func courseExerciseFromRequest(r *http.Request) (course Course, exercise Exercise, err error) {
	vars := mux.Vars(r)

	course, err = courseFromRequest(r)
	if err != nil {
		return
	}

	exercise, found := course.Exercises[vars["exercise"]]
	if !found {
		err = errors.New("exercise not found")
	}

	return
}

func apiBuildStatus(w http.ResponseWriter, r *http.Request) {
	_, exercise, err := courseExerciseFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	submissionFilename := submissionFilenameFromRequest(r)

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

func apiUpload(w http.ResponseWriter, r *http.Request) {
	course, exercise, err := courseExerciseFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	submission, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !isZIP(submission) {
		http.Error(w, "wrong content type, application/zip required.", http.StatusBadRequest)
		return
	}

	submissionFilename := submissionFilenameFromRequest(r)

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

func tokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := tokenFromRequest(r)
		if !TokenHasValidFormat(token) {
			http.Error(w, "invalid token", http.StatusBadRequest)
			return
		}

		course, err := courseFromRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		for _, user := range course.Users {
			if user == token {
				next.ServeHTTP(w, r)
				return
			}
		}

		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})
}

func Serve(addr string) {
	router := mux.NewRouter()

	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(tokenAuthMiddleware)
	apiRouter.HandleFunc("/{course}/{exercise}", apiBuildStatus).Methods(http.MethodGet)
	apiRouter.HandleFunc("/{course}/{exercise}", apiUpload).Methods(http.MethodPost)

	http.Handle("/", router)

	log.Printf("Listening on %s\n", addr)
	http.ListenAndServe(addr, nil)
}
