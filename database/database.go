package database

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/customerio/homework/serve"
)

type Customer_User struct {
	ID           int
	EMAIL        string
	FIRST_NAME   string
	LAST_NAME    string
	IP           string
	LAST_UPDATED int
	EVENT_IDS    []uint8
	EVENT_COUNT  int
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

type Database struct {
	db     *sql.DB
	db_err error
}

func (d *Database) Construct(user string, pw string, host string) {
	d.db = new(sql.DB)
	conninfo := "user=" + user + " password=" + pw + " host=" + host + " sslmode=disable"
	d.db, d.db_err = sql.Open("postgres", conninfo)

	if d.db_err != nil {
		log.Fatal(d.db_err)
	}
	dbName := "testdb"
	_, err := d.db.Exec("DROP DATABASE " + dbName + " WITH (FORCE)")
	if err != nil {
		//handle the error
		log.Fatal(err)
	}
	_, err = d.db.Exec("create database " + dbName)
	if err != nil {
		//handle the error
		log.Fatal(err)
	}

	//  kill the tables manuially. I'm not sure why these don't simply die with the database :(
	_, err = d.db.Exec("DROP TABLE IF EXISTS user_attr_updates")
	if err != nil {
		//handle the error
		log.Fatal(err)
	}
	_, err = d.db.Exec("DROP TABLE IF EXISTS event")
	if err != nil {
		//handle the error
		log.Fatal(err)
	}
	_, err = d.db.Exec("DROP TABLE IF EXISTS cust_user")
	if err != nil {
		//handle the error
		log.Fatal(err)
	}

	// user table
	_, err = d.db.Exec("CREATE TABLE IF NOT EXISTS cust_user (id BIGSERIAL PRIMARY KEY, created_at timestamp default current_timestamp, email text, first_name text, last_name text, ip text, last_updated integer);")
	if err != nil {
		//handle the error
		log.Fatal(err)
	}
	// event table
	_, err = d.db.Exec("CREATE TABLE IF NOT EXISTS event(id uuid PRIMARY KEY,type text,name text,user_id bigint references cust_user,data text,timestamp integer);")
	if err != nil {
		//handle the error
		log.Fatal(err)
	}

	// attribute update table
	_, err = d.db.Exec("CREATE TABLE IF NOT EXISTS user_attr_updates(id BIGSERIAL PRIMARY KEY, user_id bigint references cust_user, name text NOT NULL, value TEXT, created timestamp default current_timestamp);")
	if err != nil {
		//handle the error
		log.Fatal(err)
	}

	// drop them if they exist already for some reason
	_, err = d.db.Exec(`
	DROP TRIGGER IF EXISTS record_user_attribute_changes_after_insert ON cust_user;
	DROP TRIGGER IF EXISTS record_user_attribute_changes_after_update ON cust_user;
	DROP function IF EXISTS record_user_attribute_changes_function;`)
	if err != nil {
		log.Fatal(err)
	}
	// db trigger to update attribute update table when user values are changed/added
	_, err = d.db.Exec(`
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
	if err != nil {
		//handle the error
		log.Fatal(err)
	}
}

// customer queries
func (d Database) GetCustomerById(id int) (*serve.Customer, error) {
	db_user := Customer_User{
		ID:           4,
		EMAIL:        "",
		FIRST_NAME:   "",
		LAST_NAME:    "",
		IP:           "",
		LAST_UPDATED: 0,
		EVENT_IDS:    nil,
		EVENT_COUNT:  0,
	}
	err := d.db.QueryRow(`
	SELECT cust_user.id,
		COALESCE(email, '') email,
		COALESCE(first_name, '') first_name,
		COALESCE(last_name, '') last_name,
		COALESCE(ip, '') ip,
		json_build_array(event.user_id) event_ids,
		COUNT(event.id) event_count
	FROM cust_user LEFT JOIN event ON event.user_id = cust_user.id WHERE cust_user.id = $1 GROUP BY cust_user.id, event.user_id;
	`, id).Scan(&db_user.ID, &db_user.EMAIL, &db_user.FIRST_NAME, &db_user.LAST_NAME, &db_user.IP, &db_user.EVENT_IDS, &db_user.EVENT_COUNT)

	attributes := map[string]string{
		"email":        db_user.EMAIL,
		"first_name":   db_user.FIRST_NAME,
		"last_name":    db_user.LAST_NAME,
		"ip":           db_user.IP,
		"last_updated": string(db_user.LAST_UPDATED),
	}
	events := map[string]int{
		"count": db_user.EVENT_COUNT,
	}
	formatted_user := serve.Customer{

		ID:          db_user.ID,
		Attributes:  attributes,
		Events:      events,
		LastUpdated: 0,
	}
	return &formatted_user, err
}

// func (d Database) ListCustomers(page, count int) (sql.Result, error) {
// 	return query("SELECT * FROM cust_user")
// }
// func (d Database) ListCustomers(page int, count int) (*serve.Customer, error) {
// 	db_user := Customer_User{
// 		ID:           4,
// 		EMAIL:        "",
// 		FIRST_NAME:   "",
// 		LAST_NAME:    "",
// 		IP:           "",
// 		LAST_UPDATED: 0,
// 		EVENT_IDS:    nil,
// 		EVENT_COUNT:  0,
// 	}
// 	err := d.db.QueryRow(`
// 	SELECT cust_user.id,
// 		COALESCE(email, '') email,
// 		COALESCE(first_name, '') first_name,
// 		COALESCE(last_name, '') last_name,
// 		COALESCE(ip, '') ip,
// 		json_build_array(event.user_id) event_ids,
// 		COUNT(event.id) event_count
// 	FROM cust_user LEFT JOIN event ON event.user_id = cust_user.id WHERE cust_user.id = $1 GROUP BY cust_user.id, event.user_id;
// 	`, id).Scan(&db_user.ID, &db_user.EMAIL, &db_user.FIRST_NAME, &db_user.LAST_NAME, &db_user.IP, &db_user.EVENT_IDS, &db_user.EVENT_COUNT)

// 	attributes := map[string]string{
// 		"id":           string(db_user.ID),
// 		"email":        db_user.EMAIL,
// 		"first_name":   db_user.FIRST_NAME,
// 		"last_name":    db_user.LAST_NAME,
// 		"ip":           db_user.IP,
// 		"last_updated": string(db_user.LAST_UPDATED),
// 	}
// 	events := map[string]int{
// 		"count": db_user.EVENT_COUNT,
// 	}
// 	formatted_user := serve.Customer{

// 		ID:          db_user.ID,
// 		Attributes:  attributes,
// 		Events:      events,
// 		LastUpdated: 0,
// 	}
// 	return &formatted_user, err
// }

func (d Database) CreateCustomer(id int, attributes map[string]string) (*serve.Customer, error) {
	var createdId int
	// TODO this is terrible but i dont have time to parse different kinds of dates and timestamps so here we are
	attributes["last_updated"] = "0"
	last_updated, parse_err := strconv.Atoi(attributes["last_updated"])
	if parse_err != nil {
		log.Fatal(parse_err)
	}
	err := d.db.QueryRow(`
	INSERT INTO cust_user (id, email, first_name, last_name, ip, last_updated) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		id, attributes["email"], attributes["first_name"], attributes["last_name"], attributes["ip"], last_updated).Scan(&createdId)
	if err != nil {
		log.Fatal(err)
	}
	return d.GetCustomerById(createdId)
}

// func (d Database) UpdateCustomerById(c Customer_User) (sql.Result, error) {
// 	return query(fmt.Sprintf("UPDATE cust_user SET (email, first_name, last_name, ip, last_updated) VALUES (%v, %v, %v, %v, %d) WHERE id = %d",
// 		c.EMAIL, c.FIRST_NAME, c.LAST_NAME, c.IP, c.LAST_UPDATED, c.ID))
// }

func (d Database) UpdateCustomerById(id int, attributes map[string]string) (*serve.Customer, error) {
	_, err := d.db.Exec(`
	UPDATE cust_user SET (email, first_name, last_name, ip) = ($1, $2, $3, $4) WHERE cust_user.id = $5`,
		attributes["email"], attributes["first_name"], attributes["last_name"], attributes["ip"], id)

	if err != nil {
		log.Fatal(err)
	}
	return d.GetCustomerById(id)
}

// func (d Database) DeleteCustomer(id int) (sql.Result, error) {
// 	return query(fmt.Sprintf("DELETE FROM cust_user WHERE id = %d", id))
// }

// func (d Database) GetTotalCustomers() (sql.Result, error) {
// 	return query("SELECT COUNT(id) FROM cust_user")
// }

// // event queries
// func (d Database) GetEventById(id string) (sql.Result, error) {
// 	return query(fmt.Sprintf("SELECT * FROM event WHERE id = %v", id))
// }

// func (d Database) ListEvents(page, count int) (sql.Result, error) {
// 	return query("SELECT * FROM event")
// }

func (d Database) CreateEvent(e Event) (int, error) {
	var createdId int
	err := d.db.QueryRow(`
	INSERT INTO event (id, type, name, user_id, data, timestamp) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		e.ID, e.TYPE, e.NAME, e.USER_ID, e.DATA, e.TIMESTAMP).Scan(&createdId)
	if err != nil {
		log.Fatal(err)
	}
	return createdId, err
}

// // attribute change queries
// func (d Database) GetAttributeEventById(id int) (sql.Result, error) {
// 	return query(fmt.Sprintf("SELECT * FROM user_attr_updates WHERE id = %d", id))
// }

// func (d Database) GetAttributeEventsByUserId(id int) (sql.Result, error) {
// 	return query(fmt.Sprintf("SELECT * FROM user_attr_updates WHERE user_id = %d", id))
// }

// func (d Database) ListAttributeEvents() (sql.Result, error) {
// 	return query("SELECT * FROM user_attr_updates")
// }
