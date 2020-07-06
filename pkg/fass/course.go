package fass

import (
	"io/ioutil"
	"log"
	"path"
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

// Exercise represents an exercise sheet.
type Exercise struct {
	Identifier string `json:"-"`
	Path       string `json:"-"`
}

// BuildScriptPath provides the path for the build script of the exercise.
func (e Exercise) BuildScriptPath() string {
	return path.Join(e.Path, "build")
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
