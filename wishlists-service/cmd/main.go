package main

import (
	"net/http"

	"gofr.dev/pkg/gofr"
)

type Customer struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	// initialise gofr object
	app := gofr.New()

	// // get app name from configs
	// appName := app.Config.Get("APP_NAME")

	// err := migration.Migrate(appName, dbmigration.NewGorm(app.GORM()),
	// 	migrations.All(), dbmigration.UP, app.Logger)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// register metrics
	_ = app.NewCounter("Counter_Name", "Help for Counter", "label1", "label2")

	// register custom-middlware
	app.Server.UseMiddleware(customMiddleware())

	app.GET("/greet", func(ctx *gofr.Context) (interface{}, error) {
		// Get the value using the redis instance
		val := 2.0
		_ = ctx.Metric.AddCounter("Counter_Name", val, "label1", "label2")
		value, err := ctx.Redis.Get(ctx.Context, "greeting").Result()

		return value, err
	})

	app.GET("/greet", func(ctx *gofr.Context) (interface{}, error) {
		// Get the value using the redis instance
		value, err := ctx.Redis.Get(ctx.Context, "greeting").Result()

		return value, err
	})

	app.POST("/customer/{name}", func(ctx *gofr.Context) (interface{}, error) {
		name := ctx.PathParam("name")

		// Inserting a customer row in database using SQL
		_, err := ctx.DB().ExecContext(ctx, "INSERT INTO customers (name) VALUES (?)", name)

		return nil, err
	})

	app.GET("/customer", func(ctx *gofr.Context) (interface{}, error) {
		var customers []Customer

		// Getting the customer from the database using SQL
		rows, err := ctx.DB().QueryContext(ctx, "SELECT * FROM customers")
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var customer Customer
			if err := rows.Scan(&customer.ID, &customer.Name); err != nil {
				return nil, err
			}

			customers = append(customers, customer)
		}

		// return the customer
		return customers, nil
	})

	// Starts the server, it will listen on the default port 8000.
	// it can be over-ridden through configs
	app.Start()
}

func customMiddleware() func(handler http.Handler) http.Handler {
	return func(inner http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// your logic here

			// sends the request to the next middleware/request-handler
			inner.ServeHTTP(w, r)
		})
	}
}
