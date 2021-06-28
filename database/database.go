package database

import (
	"database/sql"
	"fmt"
	"log"
)

type Customer_User struct {
	ID           int
	EMAIL        string
	FIRST_NAME   string
	LAST_NAME    string
	IP           string
	LAST_UPDATED int
}

type Event struct {
	ID        string
	TYPE      string
	NAME      string
	USER_ID   int
	DATA      string
	TIMESTAMP int
}

type AttributeChangeEvent struct {
	ID        string
	USER_ID   int
	NAME      string
	VALUE     string
	TIMESTAMP int
}

type Database struct{}

var db sql.DB

func query(q string) sql.Result {
	results, err := db.Exec(q)
	if err != nil {
		//handle the error
		log.Fatal(err)
	}
	return results
}

func (d *Database) Construct(user string, pw string, host string) {
	conninfo := "user=" + user + " password=" + pw + " host=" + host + " sslmode=disable"
	db, err := sql.Open("postgres", conninfo)
	fmt.Println("database created????")

	if err != nil {
		fmt.Println("shit1")
		log.Fatal(err)
	}
	dbName := "testdb"
	_, err = db.Exec("DROP DATABASE " + dbName + " WITH (FORCE)")
	if err != nil {
		//handle the error
		log.Fatal(err)
	}
	_, err = db.Exec("create database " + dbName)
	if err != nil {
		//handle the error
		log.Fatal(err)
	}

	// user table
	query("CREATE TABLE IF NOT EXISTS cust_user (id BIGSERIAL PRIMARY KEY, created_at timestamp default current_timestamp, email text, first_name text, last_name text, ip text, last_updated timestamp default current_timestamp);")
	// event table
	query("CREATE TABLE IF NOT EXISTS event(id uuid PRIMARY KEY,type text,name text,user_id bigint references cust_user,data text,timestamp timestamp);")
	// attribute update table
	query("CREATE TABLE IF NOT EXISTS user_attr_updates(id BIGSERIAL PRIMARY KEY, user_id bigint references cust_user, name text NOT NULL, value TEXT NOT NULL, created timestamp default current_timestamp);")
	// db trigger to update attribute update table when user values are changed/added
	query(`
	CREATE FUNCTION record_user_attribute_changes_function() RETURNS trigger
    AS $$
BEGIN
    IF NEW.id IS NOT NULL THEN

    WITH attr_changes AS (
        SELECT DISTINCT ON (name) * FROM user_attr_updates WHERE user_id = NEW.id
    ),
    existing_user AS (
        SELECT *,
		-- array_to_string postgres hack filters out nulls in arrays. this ensures that we don't add events for attributes that didn't actually come through
           string_to_array(array_to_string(
            ARRAY[
                CASE WHEN NEW.email = (SELECT value FROM attr_changes WHERE name = 'email') OR NEW.email IS NULL THEN NULL ELSE 'email'::text END,
                CASE WHEN NEW.first_name = (SELECT value FROM attr_changes WHERE name = 'first_name') OR NEW.first_name IS NULL THEN NULL ELSE 'first_name'::text END,
                CASE WHEN NEW.last_name = (SELECT value FROM attr_changes WHERE name = 'last_name') OR NEW.last_name IS NULL THEN NULL ELSE 'last_name'::text END,
                CASE WHEN NEW.ip = (SELECT value FROM attr_changes WHERE name = 'ip') OR NEW.ip IS NULL THEN NULL ELSE 'ip'::text END]
                , ','
             ), ',') as keys,
           string_to_array(array_to_string(
            ARRAY[
                CASE WHEN NEW.email = (SELECT value FROM attr_changes WHERE name = 'email') THEN NULL ELSE NEW.email::text END,
                CASE WHEN NEW.first_name = (SELECT value FROM attr_changes WHERE name = 'first_name') THEN NULL ELSE NEW.first_name::text END,
                CASE WHEN NEW.last_name = (SELECT value FROM attr_changes WHERE name = 'last_name') THEN NULL ELSE NEW.last_name::text END,
                CASE WHEN NEW.ip = (SELECT value FROM attr_changes WHERE name = 'ip') THEN NULL ELSE NEW.ip::text END]
            , ','
            ), ',') as vals
        FROM cust_user WHERE id = NEW.id GROUP BY cust_user.id ORDER BY id
    ),
    new_attr_events AS (
        SELECT NEW.id,
               unnest(keys),
               unnest(vals)
        FROM existing_user
        WHERE array_length(existing_user.keys, 1) <> 0
    )
    INSERT INTO user_attr_updates (user_id, name, value)
    SELECT * FROM new_attr_events;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER record_user_attribute_changes_after_insert AFTER INSERT ON cust_user
FOR EACH ROW EXECUTE PROCEDURE record_user_attribute_changes_function();

CREATE TRIGGER record_user_attribute_changes_after_update AFTER UPDATE ON cust_user
FOR EACH ROW EXECUTE PROCEDURE record_user_attribute_changes_function();
	`)

}

// customer queries
func (d Database) GetCustomerById(id int) sql.Result {
	return query(fmt.Sprintf("SELECT * FROM cust_user WHERE id = %d", id))
}

func (d Database) ListCustomers(page, count int) sql.Result {
	return query("SELECT * FROM cust_user")
}

func (m Database) CreateCustomer(c Customer_User) sql.Result {
	return query(fmt.Sprintf("INSERT INTO cust_user (id, email, first_name, last_name, ip, last_updated) VALUES (%d, %v, %v, %v, %v, %d)",
		c.ID, c.EMAIL, c.FIRST_NAME, c.LAST_NAME, c.IP, c.LAST_UPDATED))
}

func (m Database) UpdateCustomerById(c Customer_User) sql.Result {
	return query(fmt.Sprintf("UPDATE cust_user SET (email, first_name, last_name, ip, last_updated) VALUES (%v, %v, %v, %v, %d) WHERE id = %d",
		c.EMAIL, c.FIRST_NAME, c.LAST_NAME, c.IP, c.LAST_UPDATED, c.ID))
}

func (m Database) DeleteCustomer(id int) sql.Result {
	return query(fmt.Sprintf("DELETE FROM cust_user WHERE id = %d", id))
}

func (m Database) GetTotalCustomers() sql.Result {
	return query("SELECT COUNT(id) FROM cust_user")
}

// event queries
func (d Database) GetEventById(id string) sql.Result {
	return query(fmt.Sprintf("SELECT * FROM event WHERE id = %v", id))
}

func (d Database) ListEvents(page, count int) sql.Result {
	return query("SELECT * FROM event")
}

func (m Database) CreateEvent(e Event) sql.Result {
	return query(fmt.Sprintf("INSERT INTO event (id, type, name, user_id, data, timestamp) VALUES (%v, %v, %v, %d, %v, %d)",
		e.ID, e.TYPE, e.NAME, e.USER_ID, e.DATA, e.TIMESTAMP))
}

// attribute change queries
func (d Database) GetAttributeEventById(id int) sql.Result {
	return query(fmt.Sprintf("SELECT * FROM user_attr_updates WHERE id = %d", id))
}

func (d Database) GetAttributeEventsByUserId(id int) sql.Result {
	return query(fmt.Sprintf("SELECT * FROM user_attr_updates WHERE user_id = %d", id))
}

func (d Database) ListAttributeEvents() sql.Result {
	return query("SELECT * FROM user_attr_updates")
}
