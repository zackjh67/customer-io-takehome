package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/customerio/homework/datastore"
	"github.com/customerio/homework/serve"
	"github.com/customerio/homework/stream"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("PARSING!!! @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@######################################################################################################")
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// kill the server when you get terminate signal
	go func() {
		<-sigs
		cancel()
	}()

	file, err := os.Open("data/messages.1.data")

	ds := new(datastore.Datastore)
	ds.Construct()

	ch, err := stream.Process(ctx, file)
	if err != nil {
		log.Fatal(err)
	}
	i := 0
	for rec := range ch {
		if i%1000 == 0 {
			fmt.Println(i, " lines parsed....")
		}
		// user_id is a foreign key so itll break if I don't check to make sure user_id exists
		// TODO fix that ^
		if rec.UserID != "" {
			user_id, parse_err := strconv.Atoi(rec.UserID)
			if parse_err != nil {
				log.Fatal(parse_err)
			}
			// existing_user, e := ds.DB.GetCustomerById(user_id)
			_, e := ds.DB.GetCustomerById(user_id)
			// TODO catch specific user doesnt exist error instead of this catchall
			if e != nil {
				if rec.Type == "attributes" {
					customer_attributes := map[string]string{
						"id":           rec.UserID,
						"email":        rec.Data["email"],
						"first_name":   rec.Data["first_name"],
						"last_name":    rec.Data["last_name"],
						"ip":           rec.Data["ip"],
						"last_updated": rec.Data["last_updated"],
					}

					ds.DB.CreateCustomer(
						user_id, customer_attributes,
					)
				}
			} else {
				if rec.Type == "attributes" {
					customer_attributes := map[string]string{
						"id":           rec.UserID,
						"email":        rec.Data["email"],
						"first_name":   rec.Data["first_name"],
						"last_name":    rec.Data["last_name"],
						"ip":           rec.Data["ip"],
						"last_updated": rec.Data["last_updated"],
					}
					ds.DB.UpdateCustomerById(
						user_id, customer_attributes,
					)
				}
			}
		}
		i = i + 1
	}
	if err := ctx.Err(); err != nil {
		log.Fatal(err)
	}

	if err := serve.ListenAndServe(":1323", ds); err != nil {
		log.Fatal(err)
	}
}
