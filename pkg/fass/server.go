package fass

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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

func apiListExercises(w http.ResponseWriter, r *http.Request) {
	course, err := courseFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	for _, exercise := range course.Exercises {
		fmt.Fprintln(w, exercise.Identifier)
	}
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

func apiBuildUpload(w http.ResponseWriter, r *http.Request) {
	course, exercise, err := courseExerciseFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	submission, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	submissionReader := bufio.NewReader(submission)

	if !isZIP(submissionReader) {
		http.Error(w, "wrong content type, application/zip required", http.StatusBadRequest)
		return
	}

	submissionFilename := submissionFilenameFromRequest(r)

	sha256sum, err := exercise.StoreSubmission(submissionReader, submissionFilename)
	if err != nil {
		log.Println(course.Identifier, exercise.Identifier, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Invoking build:", course.Identifier, exercise.Identifier, submissionFilename)
	go invokeBuild(exercise, submissionFilename)

	fmt.Fprintln(w, course.Identifier, exercise.Identifier, "upload successful")
	fmt.Fprintf(w, "%x\n", sha256sum)
}

func invokeBuild(exercise Exercise, submissionFilename string) {
	err := exercise.BuildSubmission(submissionFilename)
	if err != nil {
		log.Println(err)
	}
}

func apiFeedback(w http.ResponseWriter, r *http.Request) {
	course, exercise, err := courseExerciseFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	feedback, err := exercise.GetFeedback(submissionFilenameFromRequest(r))
	if os.IsNotExist(err) {
		http.Error(w, "no feedback available", http.StatusNotFound)
		return
	} else if err != nil {
		log.Println(course.Identifier, exercise.Identifier, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	io.Copy(w, feedback)
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

		if !course.IsUserAuthorized(token) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Serve starts the FASS service.
func Serve(addr string) {
	router := mux.NewRouter()

	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(tokenAuthMiddleware)
	apiRouter.HandleFunc("/{course}", apiListExercises)
	apiRouter.HandleFunc("/{course}/{exercise}/build", apiBuildStatus).Methods(http.MethodGet)
	apiRouter.HandleFunc("/{course}/{exercise}/build", apiBuildUpload).Methods(http.MethodPost)
	apiRouter.HandleFunc("/{course}/{exercise}/feedback", apiFeedback)

	http.Handle("/", router)

	log.Printf("Listening on %s\n", addr)
	http.ListenAndServe(addr, nil)
}
