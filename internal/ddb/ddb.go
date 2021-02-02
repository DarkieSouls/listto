package ddb

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/DarkieSouls/listto/internal/lists"
	"github.com/DarkieSouls/listto/internal/listtoErr"
)

const (
	table = "listto_lists"
)

type DDB struct {
	DDB *dynamodb.DynamoDB
}

func New(ddb *dynamodb.DynamoDB) *DDB {
	return &DDB{
		DDB: ddb,
	}
}

func (d *DDB) GetList(guild, lis string) (list *lists.ListtoList, lisErr *listtoErr.ListtoError) {
	defer func() {
		if lisErr != nil {
			lisErr.SetCallingMethodIfNil("GetList")
		}
	}()

	input := (&dynamodb.GetItemInput{}).SetTableName(table).SetKey(map[string]*dynamodb.AttributeValue{
		"guild": (&dynamodb.AttributeValue{}).SetS(guild),
		"name":  (&dynamodb.AttributeValue{}).SetS(lis),
	})

	output, err := d.DDB.GetItem(input)
	if err != nil {
		lisErr = listtoErr.ConvertError(err)
		return
	}

	if len(output.Item) < 1 {
		lisErr = listtoErr.ListNotFoundError(lis)
		return
	}

	list = new(lists.ListtoList)
	if err := dynamodbattribute.UnmarshalMap(output.Item, &list); err != nil {
		lisErr = listtoErr.ConvertError(err)
	}

	return
}

func (d *DDB) GetAllLists(guild, user string) (values []*lists.ListtoList, lisErr *listtoErr.ListtoError) {
	defer func() {
		if lisErr != nil {
			lisErr.SetCallingMethodIfNil("GetAllLists")
		}
	}()

	input := (&dynamodb.QueryInput{}).SetTableName(table).SetKeyConditionExpression("guild = :v1 OR guild = :v2").
		SetExpressionAttributeValues(map[string]*dynamodb.AttributeValue{":v1": (&dynamodb.AttributeValue{}).SetS(guild), ":v2": (&dynamodb.AttributeValue{}).SetS(user)})

	output, err := d.DDB.Query(input)
	if err != nil {
		lisErr = listtoErr.ConvertError(err)
		return
	}

	if len(output.Items) < 1 {
		lisErr = listtoErr.ListsNotFoundError()
		return
	}

	for _, v := range output.Items {
		lis := new(lists.ListtoList)
		if err := dynamodbattribute.UnmarshalMap(v, &lis); err != nil {
			lisErr = listtoErr.ConvertError(err)
			return
		}
		values = append(values, lis)
	}

	return
}

func (d *DDB) PutList(in interface{}) (lisErr *listtoErr.ListtoError) {
	defer func() {
		if lisErr != nil {
			lisErr.SetCallingMethodIfNil("PutList")
		}
	}()

	item, err := dynamodbattribute.MarshalMap(in)
	if err != nil {
		lisErr = listtoErr.ConvertError(err)
		return
	}

	input := (&dynamodb.PutItemInput{}).SetTableName(table).SetItem(item)

	_, err = d.DDB.PutItem(input)
	if err != nil {
		lisErr = listtoErr.ConvertError(err)
	}

	return
}

func (d *DDB) DeleteList(guild, lis string) (lisErr *listtoErr.ListtoError) {
	defer func() {
		if lisErr != nil {
			lisErr.SetCallingMethodIfNil("DeleteList")
		}
	}()

	input := (&dynamodb.DeleteItemInput{}).SetTableName(table).SetKey(map[string]*dynamodb.AttributeValue{
		"guild": (&dynamodb.AttributeValue{}).SetS(guild),
		"name":  (&dynamodb.AttributeValue{}).SetS(lis),
	})

	_, err := d.DDB.DeleteItem(input)
	if err != nil {
		lisErr = listtoErr.ConvertError(err)
	}

	return
}
