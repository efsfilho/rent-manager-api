package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var tenantNew = `{"id":"000000000000000000000000","name":"Testing Kronbua Whurdagur","cpf":"21887330020","rg":"0123456789","birth_date":1012690301}`
var tenantUpdate = `{"id":"000000000000000000000000","name":"Testing Muzgehamph Zyofa","cpf":"33734532086","rg":"9876543210","birth_date":0}`

func TestPostTenant(t *testing.T) {
	fmt.Println("POST /tenants - postTenant()")
	// Request create
	payload := strings.NewReader(tenantNew)
	req, err := http.NewRequest(http.MethodPost, "/tenants", payload)

	if err != nil {
		t.Error(err)
	}
	req.Header.Add("Content-Type", "application/json")

	// Echo instance
	e := echo.New()
	rec := httptest.NewRecorder()
	context := e.NewContext(req, rec)
	result := rec.Result()
	defer result.Body.Close()

	// Load env variables for database connection
	err = godotenv.Load()
	if err != nil {
		t.Error(err)
	}

	// Create database connection
	db.connect()
	defer db.disconnect()

	// Tests
	if assert.NoError(t, postTenant(context)) {

		var sent, received Tenant
		json.Unmarshal([]byte(tenantNew), &sent)
		json.Unmarshal(rec.Body.Bytes(), &received)

		fmt.Println("\tSENT     -", sent)
		fmt.Println("\tRECEIVED -", received)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, sent, received, "Tenant sent and received should be equal")
	}
}

func TestGetTenant(t *testing.T) {

	fmt.Println("GET /tenants - getTenant()")
	// Request create
	// body := `{"id":"000000000000000000000000","name":"Testing Kronbua Whurdagur","cpf":"21887330020","rg":"0123456789","birth_date":1012690301}`
	payload := strings.NewReader("")
	req, err := http.NewRequest(http.MethodGet, "/tenants", payload)

	if err != nil {
		t.Error(err)
	}
	req.Header.Add("Content-Type", "application/json")

	// Echo instance
	e := echo.New()
	rec := httptest.NewRecorder()
	context := e.NewContext(req, rec)
	result := rec.Result()
	defer result.Body.Close()

	// Load env variables for database connection
	err = godotenv.Load()
	if err != nil {
		t.Error(err)
	}

	// Create database connection
	db.connect()
	defer db.disconnect()

	// Tests
	if assert.NoError(t, getTenant(context)) {

		// Tenant sent in the post test
		var sent Tenant
		var received []Tenant
		json.Unmarshal([]byte(tenantNew), &sent)
		json.Unmarshal(rec.Body.Bytes(), &received)

		// Ids should be cleaned so the Id field is ignored by assert.Contains
		for i := range received {
			received[i].Id = primitive.NilObjectID
		}

		// Checks if the tenant sent through the post is returned
		assert.Contains(t, received, sent)
		fmt.Println("\tFOUND    -", sent)
		assert.Equal(t, http.StatusOK, rec.Code)
		// assert.Equal(t, sent, received, "Tenant sent and received should be equal")
	}
}

func TestPutTenant(t *testing.T) {
	fmt.Println("POST /tenants - putTenant()")
	// Load env variables for database connection
	err := godotenv.Load()
	if err != nil {
		t.Error(err)
	}

	// Create database connection
	db.connect()
	defer db.disconnect()

	// Gets all tenants stored
	var tenants []Tenant = []Tenant{}
	err = listDocuments(&tenants)
	if err != nil {
		t.Error(err)
	}

	var sent Tenant
	json.Unmarshal([]byte(tenantNew), &sent)

	// Get the id of the tenant included by TestPostTenant
	for _, t := range tenants {
		ignore := cmpopts.IgnoreFields(Tenant{}, "Id")
		if cmp.Equal(sent, t, ignore) {
			sent.Id = t.Id
			fmt.Println("\tOLD      -", t)
		}
	}

	if sent.Id.IsZero() {
		t.Errorf("Tenant used for test was not found, the test tenant should be included by the first test(TestPostTenant)")
	}

	// New tenant data
	payload := strings.NewReader(tenantUpdate)

	req, err := http.NewRequest(http.MethodPut, "/tenants", payload)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("Content-Type", "application/json")

	// Echo instance
	e := echo.New()
	rec := httptest.NewRecorder()
	context := e.NewContext(req, rec)
	context.SetParamNames("id")
	context.SetParamValues(sent.Id.Hex())
	result := rec.Result()
	defer result.Body.Close()

	var updated Tenant
	json.Unmarshal([]byte(tenantUpdate), &updated)
	fmt.Println("\tUPDATED  -", updated)

	// Tests
	if assert.NoError(t, putTenant(context)) {
		assert.Equal(t, http.StatusNoContent, rec.Code)
	}
}

func TestDeleteTenant(t *testing.T) {
	fmt.Println("DELETE /tenants - deleteTenant()")
	// Load env variables for database connection
	err := godotenv.Load()
	if err != nil {
		t.Error(err)
	}

	// Create database connection
	db.connect()
	defer db.disconnect()

	// Gets all tenants stored
	var tenants []Tenant = []Tenant{}
	err = listDocuments(&tenants)
	if err != nil {
		t.Error(err)
	}

	var updated Tenant
	json.Unmarshal([]byte(tenantUpdate), &updated)

	// Get the id of updated tenant
	for _, t := range tenants {
		ignore := cmpopts.IgnoreFields(Tenant{}, "Id")
		if cmp.Equal(updated, t, ignore) {
			updated.Id = t.Id
		}
	}

	if updated.Id.IsZero() {
		t.Errorf("Tenant used for test was not found, the test tenant should be included by the first test(TestPostTenant)")
	}

	// New tenant data
	// payload := strings.NewReader(body)

	req, err := http.NewRequest(http.MethodDelete, "/tenants", nil)
	if err != nil {
		t.Error(err)
	}
	req.Header.Add("Content-Type", "application/json")

	// Echo instance
	e := echo.New()
	rec := httptest.NewRecorder()
	context := e.NewContext(req, rec)
	context.SetParamNames("id")
	context.SetParamValues(updated.Id.Hex())
	result := rec.Result()
	defer result.Body.Close()

	fmt.Println("\tREMOVED  -", updated)

	// Tests
	if assert.NoError(t, putTenant(context)) {
		assert.Equal(t, http.StatusNoContent, rec.Code)
	}
}
