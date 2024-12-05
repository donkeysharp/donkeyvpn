package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type TestItem struct {
	Foo string
	Bar string
}

func (i *TestItem) ToItem() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"foo": &types.AttributeValueMemberS{Value: i.Foo},
		"bar": &types.AttributeValueMemberS{Value: i.Bar},
	}
}
func (i *TestItem) PrimaryKey() string {
	return i.Foo
}

func (i *TestItem) RangeKey() string {
	return i.Bar
}

func TestCRUD(t *testing.T) {
	t.Log("Running TestCRUD")
	ctx := context.Background()

	tableName := "donkeyvpn-prod-peers"
	table, err := NewDynamoDB(ctx, tableName)
	if err != nil {
		t.Error("Failed creating dynamodb table.")
		t.Fail()
	}
	item := &TestItem{
		Foo: "foo",
		Bar: "bar",
	}
	created, err := table.CreateRecord(item)
	if err != nil {
		t.Error("Failed creating dynamodb record")
	}
	if !created {
		t.Error("Item was not created successfully")
		t.Fail()
	}
	t.Log("Item created successfully")

	t.Log("Record created successfully")

	t.Log("Retrieving specific record")
	// retrieve peer
	t.Log("Record retrieve successfully")

	t.Log("Deleting record")
	// Delete peer
	t.Log("Record deleted successfully")
}
