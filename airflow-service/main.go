package main

import (
	"context"
	"fmt"

	"github.com/apache/airflow-client-go/airflow"
)

func main() {
	conf := airflow.NewConfiguration()
	conf.Host = "localhost:8080"
	conf.Scheme = "http"
	cli := airflow.NewAPIClient(conf)

	cred := airflow.BasicAuth{
		UserName: "username",
		Password: "password",
	}
	ctx := context.WithValue(context.Background(), airflow.ContextBasicAuth, cred)

	variable, _, err := cli.VariableApi.GetVariable(ctx, "foo").Execute()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(variable)
	}
}