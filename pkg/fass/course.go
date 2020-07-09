package fass

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// Course represents a course with its exercise sheets and registered users.
type Course struct {
	Identifier string `json:"-"`
	Name       string
	URL        string
	Path       string              `json:"-"`
	Exercises  map[string]Exercise `json:"-"`
	Users      []Token
}

func (c Course) IsUserAuthorized(token Token) bool {
	for _, user := range c.Users {
		if user == token {
			return true
		}
	}
	return false
}

// Exercise represents an exercise sheet.
type Exercise struct {
	Identifier string `json:"-"`
	Path       string `json:"-"`
}

func (e Exercise) StoreSubmission(submission io.Reader, filename string) error {
	os.Mkdir(e.submissionDir(), 0755)

	target, err := os.Create(path.Join(e.submissionDir(), filename))
	if err != nil {
		return err
	}
	defer target.Close()

	io.Copy(target, submission)

	return nil
}

func (e Exercise) BuildSubmission(submissionFilename string) error {
	logFile, err := os.Create(e.logFilepath(submissionFilename))
	if err != nil {
		return err
	}
	defer logFile.Close()

	cmd := exec.Command(e.buildScriptPath())
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Run()

	// store exit code
	{
		exitCode := 0
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}

		exitFile, err := os.Create(e.exitFilepath(submissionFilename))
		if err != nil {
			return err
		}
		defer exitFile.Close()

		fmt.Fprintln(exitFile, exitCode)
	}

	return err
}

func (e Exercise) GetBuildOutput(submissionFilename string) (io.Reader, error) {
	return os.Open(e.logFilepath(submissionFilename))
}

func (e Exercise) WasBuildSuccessful(submissionFilename string) bool {
	exitFile, err := os.Open(e.exitFilepath(submissionFilename))
	if err != nil {
		return false
	}

	var exitCode int
	_, err = fmt.Fscanf(exitFile, "%d", &exitCode)
	if err != nil {
		return false
	}

	return exitCode == 0
}

func (e Exercise) GetFeedback(submissionFilename string) (io.Reader, error) {
	return os.Open(e.feedbackFilepath(submissionFilename))
}

func (e Exercise) submissionDir() string {
	return path.Join(e.Path, "submissions")
}

func (e Exercise) buildScriptPath() string {
	return path.Join(e.Path, "build")
}

func (e Exercise) logFilepath(submissionFilename string) string {
	basename := strings.TrimSuffix(submissionFilename, filepath.Ext(submissionFilename))
	return path.Join(e.submissionDir(), basename+".log")
}

func (e Exercise) exitFilepath(submissionFilename string) string {
	basename := strings.TrimSuffix(submissionFilename, filepath.Ext(submissionFilename))
	return path.Join(e.submissionDir(), basename+".exit")
}

func (e Exercise) feedbackFilepath(submissionFilename string) string {
	basename := strings.TrimSuffix(submissionFilename, filepath.Ext(submissionFilename))
	return path.Join(e.submissionDir(), basename+".feedback")
}

// LoadCourses loads all courses from the given directory.
func LoadCourses(dataPath string) (courses []Course, err error) {
	dataDir, err := ioutil.ReadDir(dataPath)
	if err != nil {
		return
	}

	for _, f := range dataDir {
		if f.IsDir() {
			course, err := LoadCourse(path.Join(dataPath, f.Name()))
			if err != nil {
				log.Println("Could not load course:", f.Name())
				continue
			}

			courses = append(courses, course)
		}
	}

	return
}

// LoadCourse loads course data from disk.
func LoadCourse(coursePath string) (course Course, err error) {
	course = Course{
		Identifier: path.Base(coursePath),
		Path:       coursePath,
	}

	err = unmarshalFromFile(path.Join(coursePath, "course.json"), &course)
	if err != nil {
		return
	}

	course.Exercises, err = loadExercises(coursePath)

	return
}

func loadExercises(coursePath string) (map[string]Exercise, error) {
	courseDir, err := ioutil.ReadDir(coursePath)
	if err != nil {
		return nil, err
	}

	result := make(map[string]Exercise)
	for _, f := range courseDir {
		if f.IsDir() {
			result[f.Name()] = Exercise{
				Identifier: f.Name(),
				Path:       path.Join(coursePath, f.Name()),
			}
		}
	}

	return result, nil
}
