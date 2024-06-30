package main

import (
	"errors"
	"net/http"

	"github.com/fieldryand/goflow/v2"
)

func main() {
	options := goflow.Options{
		UIPath:       "ui/",
		ShowExamples: true,
		WithSeconds:  true,
	}
	gf := goflow.New(options)
	gf.AddJob(myJob)
	gf.Use(goflow.DefaultLogger())
	gf.Run(":8181")
}

func myJob() *goflow.Job {
	j := &goflow.Job{Name: "my-job", Schedule: "* * * * *"}
	j.Add(&goflow.Task{
		Name:       "sleep-for-one-second",
		Operator:   goflow.Command{Cmd: "sleep", Args: []string{"1"}},
		Retries:    5,
		RetryDelay: goflow.ConstantDelay{Period: 1},
	})
	j.Add(&goflow.Task{
		Name:     "get-google",
		Operator: goflow.Get{Client: &http.Client{}, URL: "https://www.google.com"},
	})
	j.Add(&goflow.Task{
		Name:     "add-two-plus-three",
		Operator: PositiveAddition{a: 2, b: 3},
	})
	j.SetDownstream(j.Task("sleep-for-one-second"), j.Task("get-google"))
	j.SetDownstream(j.Task("sleep-for-one-second"), j.Task("add-two-plus-three"))
	return j
}

type PositiveAddition struct{ a, b int }

func (o PositiveAddition) Run() (interface{}, error) {
	if o.a < 0 || o.b < 0 {
		return 0, errors.New("Can't add negative numbers")
	}
	result := o.a + o.b
	return result, nil
}
