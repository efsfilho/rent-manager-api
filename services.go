package main

import (
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Operation int8

const (
	OperationAdd    Operation = 0
	OperationUpdate Operation = 1
	OperationDelete Operation = 2
)

type Log struct {
	Operation       Operation
	Log             string
	NewValue        string
	OldValue        string
	DateOfOperation int64
	// User            int32
}

type Tenant struct {
	Id        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Active    bool               `json:"active"`
	Name      string             `json:"name"`
	Cpf       string             `json:"cpf"`
	Rg        string             `json:"rg"`
	BirthDate int64              `json:"birth_date" bson:"birth_date"`
	RentId    primitive.ObjectID `json:"rent_id" bson:"rent_id"`
	// PropertyId primitive.ObjectID `json:"property_id" bson:"property_id"`
}

type Property struct {
	Id      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Active  bool               `json:"active"`
	Name    string             `json:"name"`
	Address string             `json:"address"`
	RentId  primitive.ObjectID `json:"rent_id" bson:"rent_id"`
	Tenant  interface{}        `json:"tenant" bson:"tenant"`
	// TenantId primitive.ObjectID `json:"tenant_id" bson:"tenant_id"`
}

type Rent struct {
	Id         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Active     bool               `json:"active"`
	TenantId   primitive.ObjectID `json:"tenant_id"`
	PropertyId primitive.ObjectID `json:"property_id"`
}

type Config struct {
	documentType string
	collection   string
	dataBase     string
}

type DocsConfig struct {
	dataBase string
	configs  []Config
}

// Returns database and collection name of each document type
func (c *DocsConfig) getDocConfig(doc interface{}) (Config, error) {

	getTypeConfig := func(typeName string) (Config, error) {
		for _, c := range c.configs {
			if c.documentType == typeName {
				return c, nil
			}
		}
		return Config{}, errors.New("Tipo de documento não definidoAAAAA")
	}

	config := Config{}
	var err error = nil

	switch doc.(type) {
	case Tenant, *Tenant, *[]Tenant:
		config, err = getTypeConfig("Tenant")
	case Property, *Property, *[]Property:
		config, err = getTypeConfig("Property")
	case Rent, *Rent, *[]Rent:
		config, err = getTypeConfig("Rent")
	default:
		err = errors.New("Tipo de documento não definido")
	}

	config.dataBase = c.dataBase

	return config, err
}

func (c *DocsConfig) addConfig(newConfig Config) {
	c.configs = append(c.configs, newConfig)
}

var configs DocsConfig = DocsConfig{
	dataBase: "srv1140",
	configs: []Config{
		{
			documentType: "Tenant",
			collection:   "tenants",
		},
		{
			documentType: "Property",
			collection:   "properties",
		},
		{
			documentType: "Rent",
			collection:   "rents",
		},
	},
}

func isTenantValid(tenant Tenant) (bool, string) {
	if isValid, msg := isValidCpf(tenant.Cpf); !isValid {
		return false, msg
	}
	// TODO
	return true, ""
}

func saveDocument(ctx context.Context, doc interface{}) (primitive.ObjectID, error) {

	config, err := configs.getDocConfig(doc)

	if err != nil {
		logger(err)
		return primitive.NilObjectID, err
	}

	dataBase := db.client.Database(config.dataBase)
	coll := dataBase.Collection(config.collection)

	if ctx == nil {
		ctx = context.TODO()
	}

	result, err := coll.InsertOne(ctx, doc)

	if err != nil {
		logger(err)
		return primitive.NilObjectID, err
	}
	logger(fmt.Sprintf("Inserted %v document with _id: %v", config.documentType, result.InsertedID))

	objectId, isObjectId := result.InsertedID.(primitive.ObjectID)
	if isObjectId {
		return objectId, nil
	}

	return primitive.NilObjectID, nil
}

// Based on the type(doc interface{}), lists all documents of its collection
func listDocuments(doc interface{}, objectId primitive.ObjectID, filter interface{}) error {
	// if reflect.ValueOf(doc).Kind() != reflect.Ptr {
	// 	err := errors.New("a lista de documentos(doc interface{}) deve ser um ponteiro")
	// 	logger(err)
	// 	return err
	// }

	// docReference, isReference := doc.(*[]Tenant)
	// if !isReference {
	// 	return fmt.Errorf("not Refff")
	// }
	// fmt.Println("doc", doc)
	// fmt.Println("doc", docReference)
	config, err := configs.getDocConfig(doc)
	if err != nil {
		logger(err)
		return err
	}

	dataBase := db.client.Database(config.dataBase)
	coll := dataBase.Collection(config.collection)

	findOptions := options.Find()
	// findOptions.SetLimit(limit)
	// filter := bson.D{{Key: "rent_id", Value: "primitive.NilObjectID"}}
	if objectId != primitive.NilObjectID {
		filter := bson.M{"_id": objectId}
		err := coll.FindOne(context.TODO(), filter).Decode(doc)
		// fmt.Println("id", objectId, result)
		// err = result.Err()
		// result.Decode(doc)
		// .Decode(&doc)
		if err != nil {
			// logger(err)
			return err
		}
		// if err != nil {
		// 	if err == mongo.ErrNoDocuments {
		// 		// This error means your query did not match any documents.
		// 		fmt.Println("no document")
		// 		return err
		// 	}
		// 	// panic(err)
		// }
	} else {
		isPointer := reflect.ValueOf(doc).Kind() == reflect.Ptr
		isSlice := reflect.ValueOf(doc).Elem().Kind() == reflect.Slice
		if isPointer && !isSlice {
			err := fmt.Errorf("doc should point to a slice")
			logger(err)
			return err
		}

		if filter == nil {
			filter = bson.D{{}}
		}
		fmt.Println("filter ", filter)
		result, err := coll.Find(context.TODO(), filter, findOptions)
		if err != nil {
			return err
		}
		// result.Decode(&doc)
		err = result.All(context.TODO(), doc)
		if err != nil {
			return err
		}
	}

	// if err != nil {
	// 	logger(err)
	// 	return err
	// }

	// if err = result.All(context.TODO(), doc); err != nil {
	// 	logger(err)
	// 	return err
	// }

	return nil
}

func updateDocument(objectId primitive.ObjectID, doc interface{}) (int64, error) {

	config, err := configs.getDocConfig(doc)
	if err != nil {
		logger(err)
		return 0, err
	}

	filter := bson.M{"_id": objectId}
	// update := bson.A{"$set", doc}
	update := bson.D{{"$set", doc}}

	dataBase := db.client.Database(config.dataBase)
	coll := dataBase.Collection(config.collection)
	result, err := coll.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		logger(err)
		return 0, err
	}

	return result.ModifiedCount, nil
}

func removeDocument(objectId primitive.ObjectID, doc interface{}) (int64, error) {

	config, err := configs.getDocConfig(doc)
	if err != nil {
		logger(err)
		return 0, err
	}

	filter := bson.M{"_id": objectId}
	dataBase := db.client.Database(config.dataBase)
	coll := dataBase.Collection(config.collection)
	result, err := coll.DeleteOne(context.TODO(), filter)

	if err != nil {
		logger(err)
		return 0, err
	}
	return result.DeletedCount, nil
}

func saveTenant(tenant Tenant) error {

	// Checks if there is property x tenant relationship
	// isValidId := primitive.IsValidObjectID(tenant.PropertyId.Hex())
	// if isValidId && !tenant.PropertyId.IsZero() {

	// 	var properties Property = Property{}
	// 	err := listDocuments(&properties, tenant.PropertyId)

	// 	// If no document is found, the returned error is going to be mongo.ErrNoDocuments
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	tenant.Active = true
	_, err := saveDocument(nil, tenant)

	if err != nil {
		logger(err)
		return err
	}
	return nil
}

func saveProperty(property Property) error {

	// Checks if there is property x tenant relationship
	// isValidId := primitive.IsValidObjectID(property.TenantId.Hex())
	// if isValidId && !property.TenantId.IsZero() {

	// 	var tenant Tenant = Tenant{}
	// 	err := listDocuments(&tenant, property.TenantId)

	// 	// If no document is found, the returned error is going to be mongo.ErrNoDocuments
	// 	if err != nil {
	// 		return err
	// 	}
	// }
	property.Active = true
	_, err := saveDocument(nil, property)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			logger(err)
			// TODO check err
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "Inquilino não encontrado!")
		} else {
			logger(err)
			return err
		}
	}

	return nil
}

func listProperties() ([]Property, error) {
	var properties []Property = []Property{}
	err := listDocuments(&properties, primitive.NilObjectID, nil)

	if err != nil {
		return nil, err
	}

	// Fills tenant field if any data is available, otherwise it should be null
	for i, a := range properties {
		if !a.RentId.IsZero() {
			tenant := Tenant{}
			// Checks if there are any rent (tenant/property relation)
			err := listDocuments(&tenant, a.RentId, nil)
			if err != nil {
				logger(fmt.Sprintf("An error occurred when looking for rent document: %v ", a.RentId.String()))
				logger(err)
				break
			}
			properties[i].Tenant = &tenant
		}
	}

	return properties, nil
}

func saveRent(rent Rent) error {

	return nil
}

// func saveLog(op Operation, oldValue string, msg string) {

// 	reg := Log{}
// 	reg.Operation = op

// 	coll := db.client.Database("srv1140").Collection("logs")
// 	result, err := coll.InsertOne(context.TODO(), tenant)
// 	fmt.Println("SS", a)
// }
