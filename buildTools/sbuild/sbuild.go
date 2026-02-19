package sbuild

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// set buildtools working dir
var WORKING_DIR = "./"

type component struct {
	Name          string   `json:"name"`
	BuildCommand  string   `json:"buildCommand"`
	Dependencies  []string `json:"dependencies"`
	ComponentPath string
	IsBuilt       bool
}

func indexSingleComponents(d string) (component, error) {
	f, e := os.ReadFile(d + "/component.json")

	if e != nil {
		return component{}, e
	}

	var res = component{}
	json.Unmarshal(f, &res)

	res.ComponentPath = d + "/"
	res.IsBuilt = false

	return res, nil
}

func IndexingComponents() []component {
	var res = []component{}

	dirs, e := os.ReadDir(WORKING_DIR + ".")

	if e != nil {
		fmt.Println("somethings wrong!")
		return res
	}

	for _, entry := range dirs {
		if entry.IsDir() {
			c, e := indexSingleComponents(WORKING_DIR + entry.Name())

			if e == nil {

				res = append(res, c)
			}
		}
	}

	return res
}

// for faster searching
func MappingComponents(cs []component) map[string]component {
	var res = make(map[string]component)

	for i, v := range cs {
		res[v.Name] = cs[i]
	}

	return res
}

func BuildComponent(c component) (string, error) {
	if c.IsBuilt {
		return "", nil
	}

	fmt.Println("Build component '" + c.Name + "'")

	cmd := exec.Command("sh", "-c", c.BuildCommand)
	cmd.Dir = c.ComponentPath

	output, e := cmd.CombinedOutput()

	if e != nil {
		fmt.Println("Error: Build component '" + c.Name + "' fail!\n-----LOG-----")
		fmt.Println(output)
		return string(output), e
	}

	fmt.Println("Component '" + c.Name + "' build successful!")

	return string(output), nil
}

type ShipInfo struct {
	Name         string   `json:"name"`
	Dependencies []string `json:"dependencies"`
}

func Ship() error {
	var info ShipInfo

	b, e := os.ReadFile(WORKING_DIR + "/ship.json")

	if e != nil {
		fmt.Println("Error: ship.json can't be opened!")
		return e
	}

	e = json.Unmarshal(b, &info)

	if e != nil {
		fmt.Println("Error: ship.json has bad format!")
		return e
	}

	// build neccessary components
	fmt.Println("Indexing component(s)...")
	cs := IndexingComponents()
	csm := MappingComponents(cs)

	for _, v := range info.Dependencies {
		c, ok := csm[v]

		if !ok {
			fmt.Println("Error: no such component '" + v + "'!")
			return errors.New("NO_COMPONENT")
		}

		if c.IsBuilt {
			continue
		}

		// build dependencies
		for _, v := range c.Dependencies {
			sc, ok := csm[v]

			if !ok {
				fmt.Println("Error: no such component '" + v + "'!")
				return errors.New("NO_COMPONENT")
			}

			BuildComponent(sc)
			sc.IsBuilt = true
			csm[v] = sc
		}

		_, e := BuildComponent(c)
		if e != nil {

			return e
		}

		c.IsBuilt = true
		csm[v] = c
	}

	return nil
}
