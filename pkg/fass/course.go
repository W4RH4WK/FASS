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
	Exercises  []Exercise `json:"-"`
	Users      []Token
}

// Exercise represents an exercise sheet.
type Exercise struct {
	Identifier string `json:"-"`
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
	course.Identifier = path.Base(coursePath)

	err = unmarshalFromFile(path.Join(coursePath, "course.json"), &course)
	if err != nil {
		return
	}

	course.Exercises, err = loadExercises(coursePath)

	return
}

func loadExercises(coursePath string) (exercises []Exercise, err error) {
	courseDir, err := ioutil.ReadDir(coursePath)
	if err != nil {
		return
	}

	for _, f := range courseDir {
		if f.IsDir() {
			exercises = append(exercises, Exercise{Identifier: f.Name()})
		}
	}

	return
}
