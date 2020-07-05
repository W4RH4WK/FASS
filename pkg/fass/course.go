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
			course, err := LoadCourse(f.Name())
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
func LoadCourse(identifier string) (course Course, err error) {
	course.Identifier = identifier

	err = unmarshalFromFile(path.Join(identifier, "course.json"), &course)
	if err != nil {
		return
	}

	course.Exercises, err = loadExercises(identifier)

	return
}

func loadExercises(courseIdentifier string) (exercises []Exercise, err error) {
	courseDir, err := ioutil.ReadDir(courseIdentifier)
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
