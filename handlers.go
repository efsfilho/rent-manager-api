package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func postTenant(c echo.Context) error {
	tenant := Tenant{}

	if err := c.Bind(&tenant); err != nil {
		logger(err)
		return err
	}

	if isValid, msg := isTenantValid(tenant); !isValid {
		logger(msg)
		return echo.NewHTTPError(http.StatusUnprocessableEntity, msg)
	}

	err := saveDocument(tenant)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusCreated, tenant)
}

func getTenant(c echo.Context) error {

	var tenants []Tenant = []Tenant{}
	err := listDocuments(&tenants)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return c.JSON(http.StatusOK, tenants)
}

func putTenant(c echo.Context) error {

	// Get Object id from params
	id := c.Param("id")

	if !primitive.IsValidObjectID(id) {
		msg := "Id do objeto inválido"
		return echo.NewHTTPError(http.StatusUnprocessableEntity, msg)
	}

	tenant := Tenant{}

	if err := c.Bind(&tenant); err != nil {
		logger(err)
		return err
	}

	// Clear ObjectId if its not null
	if !tenant.Id.IsZero() {
		tenant.Id = primitive.NilObjectID
	}

	// if isValid, msg := isTenantValid(tenant); !isValid {
	// 	log.Println(msg)
	// 	return echo.NewHTTPError(http.StatusUnprocessableEntity, msg)
	// }

	objectId, _ := primitive.ObjectIDFromHex(id)

	result, err := updateDocument(objectId, tenant)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	} else {
		if result > 0 {
			return c.JSON(http.StatusNoContent, "Registro atualizado")
		} else {
			return echo.NewHTTPError(http.StatusNotFound, "Registro não encontrado")
		}
	}
}

func deleteTenant(c echo.Context) error {

	// Get Object id from params
	id := c.Param("id")

	if !primitive.IsValidObjectID(id) {
		msg := "Id do objeto inválido"
		return echo.NewHTTPError(http.StatusUnprocessableEntity, msg)
	}

	objectId, _ := primitive.ObjectIDFromHex(id)
	result, err := removeDocument(objectId, Tenant{})

	// if err != nil {
	// 	return echo.NewHTTPError(http.StatusNotFound, "Registro não encontrado", err)
	// } else {
	// 	return c.JSON(http.StatusOK, "Registro removido")
	// }

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	} else {
		if result > 0 {
			return c.JSON(http.StatusNoContent, "Registro atualizado")
		} else {
			return echo.NewHTTPError(http.StatusNotFound, "Registro não encontrado")
		}
	}
}
