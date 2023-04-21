package main

import (
	"context"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	Id primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	// Id primitive.ObjectID `bson:"_id,omitempty"`
	// Id int32 `json:"id"`
	// Reg       string `json:"reg"`
	Active     bool
	Name       string `json:"name"`
	Cpf        string `json:"cpf"`
	Rg         string `json:"rg"`
	BirthDate  int64  `json:"birth_date"`
	PropertyId primitive.ObjectID
}

type Property struct {
	Id      primitive.ObjectID
	Name    string
	Address string
}

type Config struct {
	DocumentType string
	DataBase     string
	Collection   string
}

// Configuration related to documents typed as Tenant
var tenantConfig = Config{
	"Tenant",
	"srv1140", // Database name
	"tenants", // Database collection name
}

// Returns database and collection name of each document type
func getDocConfig(obj interface{}) (Config, error) {
	var config Config

	switch obj.(type) {
	case Tenant, *[]Tenant:
		config = tenantConfig
	case int:
		fmt.Println("int type")
	default:
		return Config{}, errors.Errorf("Tipo de documento não definido")
	}
	return config, nil
}

func isTenantValid(tenant Tenant) (bool, string) {
	if isValid, msg := isValidCpf(tenant.Cpf); !isValid {
		return false, msg
	}
	// TODO
	return true, ""
}

func saveDocument(obj interface{}) error {

	config, err := getDocConfig(obj)
	if err != nil {
		logger(err)
		return err
	}

	dataBase := db.client.Database(config.DataBase)
	coll := dataBase.Collection(config.Collection)
	result, err := coll.InsertOne(context.TODO(), obj)

	if err != nil {
		logger(err)
		return err
	}
	logger(fmt.Sprintf("Inserted %v document with _id: %v", config.DocumentType, result.InsertedID))
	return nil
}

// Based on the type(obj interface{}), lists all documents of its collection
func listDocuments(obj interface{}) error {

	if reflect.ValueOf(obj).Kind() != reflect.Ptr {
		err := fmt.Errorf("a lista de documentos(doc interface{}) deve ser uma referencia")
		logger(err)
		return err
	}

	// docReference, isReference := doc.(*[]Tenant)
	// if !isReference {
	// 	return fmt.Errorf("not Refff")
	// }
	// fmt.Println("doc", doc)
	// fmt.Println("doc", docReference)
	config, err := getDocConfig(obj)
	if err != nil {
		logger(err)
		return err
	}

	dataBase := db.client.Database(config.DataBase)
	coll := dataBase.Collection(config.Collection)

	findOptions := options.Find()
	// findOptions.SetLimit(2)
	result, err := coll.Find(context.TODO(), bson.D{{}}, findOptions)

	if err != nil {
		logger(err)
		return err
	}

	if err = result.All(context.TODO(), obj); err != nil {
		logger(err)
		return err
	}

	return nil
}

func updateDocument(objectId primitive.ObjectID, obj interface{}) (int64, error) {

	config, err := getDocConfig(obj)
	if err != nil {
		logger(err)
		return 0, err
	}

	// ttype := reflect.ValueOf(ref).Type()
	// fmt.Println("doc", ttype)
	// docReference, isReference := doc.(*[]Tenant)
	// if !isReference {
	// 	return fmt.Errorf("not Refff")
	// }

	filter := bson.M{"_id": objectId}
	// a := {"$set", ref}
	update := bson.D{{"$set", obj}}

	dataBase := db.client.Database(config.DataBase)
	coll := dataBase.Collection(config.Collection)
	result, err := coll.UpdateOne(context.TODO(), filter, update)

	if err != nil {
		logger(err)
		return 0, err
	}

	return result.ModifiedCount, nil
}

func removeDocument(objectId primitive.ObjectID, obj interface{}) (int64, error) {

	config, err := getDocConfig(obj)
	if err != nil {
		logger(err)
		return 0, err
	}

	filter := bson.M{"_id": objectId}
	dataBase := db.client.Database(config.DataBase)
	coll := dataBase.Collection(config.Collection)
	result, err := coll.DeleteOne(context.TODO(), filter)

	if err != nil {
		logger(err)
		return 0, err
	}
	return result.DeletedCount, nil
}

// func saveLog(op Operation, oldValue string, msg string) {

// 	reg := Log{}
// 	reg.Operation = op

// 	coll := db.client.Database("srv1140").Collection("logs")
// 	result, err := coll.InsertOne(context.TODO(), tenant)
// 	fmt.Println("SS", a)
// }
