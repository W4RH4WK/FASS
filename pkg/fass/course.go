package fass

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
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

func (c Course) dataFilepath() string {
	return path.Join(c.Path, "course.json")
}

func (c Course) Store() error {
	err := os.MkdirAll(c.Identifier, 0755)
	if err != nil {
		return err
	}

	return marshalToFile(c.dataFilepath(), c)
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
	Deadline   time.Time
}

func (e Exercise) dataFilepath() string {
	return path.Join(e.Path, "exercise.json")
}

func (e Exercise) Store() error {
	err := os.MkdirAll(e.Path, 0755)
	if err != nil {
		return err
	}

	return marshalToFile(e.dataFilepath(), e)
}

func (e Exercise) StoreSubmission(submission io.Reader, filename string) ([]byte, error) {
	os.MkdirAll(e.submissionDir(), 0755)

	target, err := os.Create(path.Join(e.submissionDir(), filename))
	if err != nil {
		return nil, err
	}
	defer target.Close()

	tee := io.TeeReader(submission, target)

	hasher := sha256.New()
	io.Copy(hasher, tee)

	return hasher.Sum(nil), nil
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

	err = unmarshalFromFile(course.dataFilepath(), &course)
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
			exercise, err := loadExercise(path.Join(coursePath, f.Name()))
			if err == nil {
				result[exercise.Identifier] = exercise
			}
		}
	}

	return result, nil
}

func loadExercise(exercisePath string) (Exercise, error) {
	exercise := Exercise{
		Identifier: path.Base(exercisePath),
		Path:       exercisePath,
	}

	err := unmarshalFromFile(exercise.dataFilepath(), &exercise)
	return exercise, err
}

func BuildSubmission(course Course, exercise Exercise, user Token, submissionFilename string) error {
	logFile, err := os.Create(exercise.logFilepath(submissionFilename))
	if err != nil {
		return err
	}
	defer logFile.Close()

	submissionFilepath := path.Join(exercise.submissionDir(), submissionFilename)
	submissionFilepath, err = filepath.Abs(submissionFilepath)
	if err != nil {
		return err
	}

	cmd := exec.Command(exercise.buildScriptPath())
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Env = []string{
		"FASS_COURSE=" + course.Identifier,
		"FASS_EXERCISE=" + exercise.Identifier,
		"FASS_USER=" + user,
		"FASS_SUBMISSION=" + submissionFilepath,
	}

	err = cmd.Run()

	// store exit code
	{
		exitCode := 0
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}

		exitFile, err := os.Create(exercise.exitFilepath(submissionFilename))
		if err != nil {
			return err
		}
		defer exitFile.Close()

		fmt.Fprintln(exitFile, exitCode)
	}

	return err
}
