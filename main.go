package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

type Contact struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Number string `json:"number"`
}

type Response struct {
	Contacts []Contact `json:"contacts"`
}

func main() {
	e := echo.New()
	e.GET("/contacts", getAllContacts)
	e.GET("/contacts/:id", getContact)
	e.POST("/contacts", addContact)
	e.PUT("/contacts/:id", updateContact)
	e.DELETE("/contacts/:id", deleteContact)
	e.Logger.Fatal(e.Start(":8000"))
}

// PostgreSQL veritabanına bağlanma
func dbConnect() (*sql.DB, error) {
	db, err := sql.Open("postgres", "postgresql://postgres:12345678@localhost/contacts?sslmode=disable")
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Tüm kişileri listeleme endpoint'i
func getAllContacts(c echo.Context) error {
	db, err := dbConnect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM contacts")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	contacts := []Contact{}
	for rows.Next() {
		contact := Contact{}
		err := rows.Scan(&contact.ID, &contact.Name, &contact.Number)
		if err != nil {
			log.Fatal(err)
			continue
		}
		contacts = append(contacts, contact)
	}

	response := Response{
		Contacts: contacts,
	}

	return c.JSON(http.StatusOK, response)
}

// Belirli bir kişiyi getirme endpoint'i
func getContact(c echo.Context) error {
	db, err := dbConnect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	id := c.Param("id")

	contact := Contact{}
	err = db.QueryRow("SELECT * FROM contacts WHERE id = $1", id).Scan(&contact.ID, &contact.Name, &contact.Number)
	if err != nil {
		log.Fatal(err)
		return c.NoContent(http.StatusNotFound)
	}

	response := Response{
		Contacts: []Contact{contact},
	}

	return c.JSON(http.StatusOK, response)
}

// Kişi ekleme endpoint'i
func addContact(c echo.Context) error {
	db, err := dbConnect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	contact := new(Contact)
	if err := c.Bind(contact); err != nil {
		return err
	}

	// UUID oluşturma
	id := uuid.New().String()
	contact.ID = id

	// Kişiyi veritabanına ekleme
	_, err = db.Exec("INSERT INTO contacts (id, name, number) VALUES ($1, $2, $3)", contact.ID, contact.Name, contact.Number)
	if err != nil {
		log.Fatal(err)
	}

	response := struct {
		Message string  `json:"message"`
		Contact Contact `json:"contact"`
	}{
		Message: "Contact successfully added",
		Contact: *contact,
	}

	return c.JSON(http.StatusCreated, response)
}

// Kişi güncelleme endpoint'i
func updateContact(c echo.Context) error {
	db, err := dbConnect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	id := c.Param("id")

	contact := new(Contact)
	if err := c.Bind(contact); err != nil {
		return err
	}

	// Kişiyi veritabanında güncelleme
	_, err = db.Exec("UPDATE contacts SET name = $1, number = $2 WHERE id = $3", contact.Name, contact.Number, id)
	if err != nil {
		log.Fatal(err)
		return c.NoContent(http.StatusNotFound)
	}

	contact.ID = id
	response := Response{
		Contacts: []Contact{*contact},
	}

	return c.JSON(http.StatusOK, response)
}

// Kişi silme endpoint'i
func deleteContact(c echo.Context) error {
	db, err := dbConnect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	id := c.Param("id")

	// Kişiyi veritabanından silme
	_, err = db.Exec("DELETE FROM contacts WHERE id = $1", id)
	if err != nil {
		log.Fatal(err)
		return c.NoContent(http.StatusNotFound)
	}

	response := struct {
		Message string `json:"message"`
	}{
		Message: "Contact successfully deleted.",
	}

	return c.JSON(http.StatusOK, response)
}
