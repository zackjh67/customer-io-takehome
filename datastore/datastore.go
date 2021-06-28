package datastore

import (
	"errors"

	"github.com/customerio/homework/database"
	"github.com/customerio/homework/serve"
)

type Datastore struct {
	DB *database.Database
}

func (d *Datastore) Construct() {
	d.DB = new(database.Database)
	d.DB.Construct("postgres", "password1", "localhost")
}

func (d Datastore) Get(id int) (*serve.Customer, error) {
	return d.DB.GetCustomerById(id)
}

func (d Datastore) List(page, count int) ([]*serve.Customer, error) {
	return nil, errors.New("unimplemented")
}

func (d Datastore) Create(id int, attributes map[string]string) (*serve.Customer, error) {
	return d.DB.CreateCustomer(id, attributes)
}

func (d Datastore) Update(id int, attributes map[string]string) (*serve.Customer, error) {
	return nil, errors.New("unimplemented")
}

func (d Datastore) Delete(id int) error {
	return errors.New("unimplemented")
}

func (d Datastore) TotalCustomers() (int, error) {
	return 0, errors.New("unimplemented")
}
