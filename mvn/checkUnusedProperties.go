package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
)

type Project struct {
	XMLName       xml.Name        `xml:"project"`
	Properties    PropertiesArray `xml:"properties,omitempty"`
	Modules       ModulesArray    `xml:"modules,omitempty"`
	Resources     []Resource      `xml:"build>resources>resource,omitempty"`
	TestResources []Resource      `xml:"build>testResources>testResource,omitempty"`
	Profiles      []Profile       `xml:"profiles>profile"`
}

type Profile struct {
	Properties    PropertiesArray `xml:"properties,omitempty"`
	Resources     []Resource      `xml:"build>resources>resource,omitempty"`
	TestResources []Resource      `xml:"build>testResources>testResource,omitempty"`
}

type Resource struct {
	Directory string   `xml:"directory,omitempty"`
	Filtering bool     `xml:"filtering,omitempty"`
	Includes  []string `xml:"includes>include,omitempty"`
	// Obvio la sección excludes porque no se utiliza en los proyectos de amiga
}

type PropertiesArray struct {
	PropertyList []Property `xml:",any"`
}

type ModulesArray struct {
	ModulesList []Module `xml:",any"`
}

type Module struct {
	Name string `xml:",chardata"`
}

type Property struct {
	XMLName xml.Name `xml:""`
	Value   string   `xml:",chardata"`
}

const (
	DIRECTORY_WORKERS = 4
)

var pendingDirectories int32

func open(directory string) *os.File {
	dir, err := os.Open(directory)
	if err != nil {
		panic(err)
	}
	return dir
}

func getFileContents(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		panic(err)
	}
	defer file.Close()

	result, _ := ioutil.ReadAll(file)

	return result
}

func isPropertyOnContent(content string, property string) bool {
	return strings.Contains(content, "${"+property+"}")
}

func checkPendingPropertiesOnFile(file string, pendingProperties []string, parentPendingProperties []string) ([]string, []string) {
	resultPending := make([]string, 0)
	resultParentPending := make([]string, 0)

	b := getFileContents(file)
	content := string(b)

	for _, property := range pendingProperties {
		if !isPropertyOnContent(content, property) {
			resultPending = append(resultPending, property)
		}
	}

	for _, property := range parentPendingProperties {
		if !isPropertyOnContent(content, property) {
			resultParentPending = append(resultParentPending, property)
		}
	}

	return resultPending, resultParentPending
}

func checkPendingPropertiesOnResource(pendingProperties []string, dir string, resource *Resource, parentPendingProperties []string) ([]string, []string) {
	resultPending := make([]string, 0)
	resultParentPending := make([]string, 0)

	for _, s := range pendingProperties {
		resultPending = append(resultPending, s)
	}
	for _, s := range parentPendingProperties {
		resultParentPending = append(resultParentPending, s)
	}

	for _, includedFile := range resource.Includes {
		resultPending, resultParentPending = checkPendingPropertiesOnFile(filepath.Join(dir, includedFile), resultPending, resultParentPending)
	}

	return resultPending, resultParentPending
}

func processPom(dir string, file string, parentPendingProperties []string) []string {
	var project Project

	pomFile := getFileContents(filepath.Join(dir, file))

	if err := xml.Unmarshal(pomFile, &project); err != nil {
		fmt.Println(err)
		panic(err)
	}

	t := string(pomFile)

	pendingProperties := make([]string, 0, len(project.Properties.PropertyList))
	result := make([]string, 0)

	for _, property := range parentPendingProperties {
		if !isPropertyOnContent(t, property) {
			result = append(result, property)
		}
	}

	for _, property := range project.Properties.PropertyList {
		if !isPropertyOnContent(t, property.XMLName.Local) {
			pendingProperties = append(pendingProperties, property.XMLName.Local)
		}
	}

	for _, profile := range project.Profiles {
		for _, property := range profile.Properties.PropertyList {
			if !isPropertyOnContent(t, property.XMLName.Local) {
				pendingProperties = append(pendingProperties, property.XMLName.Local)
			}
		}
	}

	if len(pendingProperties) > 0 || len(result) > 0 {
		for _, resource := range project.Resources {
			if resource.Filtering && (len(pendingProperties) > 0 || len(result) > 0) {
				pendingProperties, result = checkPendingPropertiesOnResource(pendingProperties, filepath.Join(dir, resource.Directory), &resource, result)
			}
		}
	}

	if len(pendingProperties) > 0 || len(result) > 0 {
		for _, profile := range project.Profiles {
			for _, resource := range profile.Resources {
				if resource.Filtering && (len(pendingProperties) > 0 || len(result) > 0) {
					pendingProperties, result = checkPendingPropertiesOnResource(pendingProperties, filepath.Join(dir, resource.Directory), &resource, result)
				}
			}
		}
	}

	if len(pendingProperties) > 0 || len(result) > 0 {
		for _, resource := range project.TestResources {
			if resource.Filtering && (len(pendingProperties) > 0 || len(result) > 0) {
				pendingProperties, result = checkPendingPropertiesOnResource(pendingProperties, filepath.Join(dir, resource.Directory), &resource, result)
			}
		}
	}

	if len(pendingProperties) > 0 || len(result) > 0 {
		for _, profile := range project.Profiles {
			for _, resource := range profile.TestResources {
				if resource.Filtering && (len(pendingProperties) > 0 || len(result) > 0) {
					pendingProperties, result = checkPendingPropertiesOnResource(pendingProperties, filepath.Join(dir, resource.Directory), &resource, result)
				}
			}
		}
	}

	if len(project.Modules.ModulesList) > 0 {
		for _, module := range project.Modules.ModulesList {
			pendingProperties = processPom(filepath.Join(dir, module.Name), "pom.xml", pendingProperties)
		}

		for _, property := range pendingProperties {
			fmt.Printf("Property: %s not used in file %s\n", property, filepath.Join(dir, file))
		}
	} else {
		if len(pendingProperties) > 0 {
			for _, property := range pendingProperties {
				fmt.Printf("Property: %s not used in file %s\n", property, filepath.Join(dir, file))
			}
		}
	}

	return result /* parentPendingProperties not used */
}

/*
Esta función busca módulos (que no sub-módulos en los sub-directorios).
Cuando encuentra un pom.xml en un directorio detiene el procesamiento de esa rama.
*/
func processDirectory(c chan *os.File, end chan int) {

	for dir := range c {
		// Comprobar si hay un fichero pom.xml en el directorio
		// Si lo hay, procesarlo.
		// Si no lo hay, procesar todos sus subdirectorios.
		if fileList, err := dir.Readdir(-1); err == nil {
			var found bool = false
			for _, f := range fileList {
				if !f.IsDir() && (f.Name() == "pom.xml") {
					processPom(dir.Name(), f.Name(), []string{})
					found = true
				}
			}

			if !found {
				for _, f := range fileList {
					if f.IsDir() {
						d := open(filepath.Join(dir.Name(), f.Name()))

						atomic.AddInt32(&pendingDirectories, 1)
						c <- d
					}
				}
			}
		} else {
			panic(err)
		}

		atomic.AddInt32(&pendingDirectories, -1)
		if atomic.CompareAndSwapInt32(&pendingDirectories, 0, 0) {
			close(c)
		}
	}
	end <- 1
}

func main() {
	dir := open(".")

	c := make(chan *os.File, 1000)
	end := make(chan int)

	for i := 0; i < DIRECTORY_WORKERS; i++ {
		go processDirectory(c, end)
	}

	atomic.AddInt32(&pendingDirectories, 1)
	c <- dir

	for i := 0; i < DIRECTORY_WORKERS; i++ {
		<-end
	}
}
