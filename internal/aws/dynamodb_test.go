//go:build integration

package aws

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Scratch table used by the integration test. It needs a string hash key
// named "Id" and no range key.
const integrationTableEnvVar = "DONKEYVPN_IT_TABLE"

type TestItem struct {
	Id  string
	Bar string
}

func (i *TestItem) ToItem() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"Id":  &types.AttributeValueMemberS{Value: i.Id},
		"Bar": &types.AttributeValueMemberS{Value: i.Bar},
	}
}

func (i *TestItem) PrimaryKey() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"Id": &types.AttributeValueMemberS{Value: i.Id},
	}
}

func (i *TestItem) RangeKey() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"Bar": &types.AttributeValueMemberS{Value: i.Bar},
	}
}

func (i *TestItem) String() string {
	return fmt.Sprintf("Id: %v, Bar: %v", i.Id, i.Bar)
}

func (i *TestItem) UpdateExpression() (*expression.Expression, error) {
	update := expression.Set(expression.Name("Bar"), expression.Value(i.Bar))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return nil, err
	}
	return &expr, nil
}

var (
	_ DynamoDBItem = (*TestItem)(nil)
	_ UpdateItem   = (*TestItem)(nil)
)

func stringAttribute(t *testing.T, record map[string]types.AttributeValue, name string) string {
	t.Helper()
	value, ok := record[name].(*types.AttributeValueMemberS)
	if !ok {
		t.Fatalf("Attribute %v is missing or is not a string: %v", name, record[name])
	}
	return value.Value
}

func TestCRUD(t *testing.T) {
	tableName := os.Getenv(integrationTableEnvVar)
	if tableName == "" {
		t.Skipf("%v is not set, skipping DynamoDB integration test", integrationTableEnvVar)
	}

	table, err := NewDynamoDB(context.Background(), tableName)
	if err != nil {
		t.Fatalf("Failed creating dynamodb table: %v", err)
	}

	item := &TestItem{
		Id:  fmt.Sprintf("it-%d", time.Now().UnixNano()),
		Bar: "bar",
	}
	t.Cleanup(func() {
		if err := table.DeleteRecord(item); err != nil {
			t.Errorf("Failed cleaning up record %v: %v", item.Id, err)
		}
	})

	// Create
	created, err := table.CreateRecord(item)
	if err != nil {
		t.Fatalf("Failed creating dynamodb record: %v", err)
	}
	if !created {
		t.Fatal("Item was not created successfully")
	}

	// Get
	record, err := table.GetRecord(item)
	if err != nil {
		t.Fatalf("Failed retrieving dynamodb record: %v", err)
	}
	if record == nil {
		t.Fatalf("Record %v was not found after creating it", item.Id)
	}
	if got := stringAttribute(t, record, "Bar"); got != "bar" {
		t.Errorf("Unexpected Bar after create: got %q, want %q", got, "bar")
	}

	// Update
	item.Bar = "baz"
	if err := table.UpdateRecord(item); err != nil {
		t.Fatalf("Failed updating dynamodb record: %v", err)
	}
	record, err = table.GetRecord(item)
	if err != nil {
		t.Fatalf("Failed retrieving dynamodb record after update: %v", err)
	}
	if record == nil {
		t.Fatalf("Record %v was not found after updating it", item.Id)
	}
	if got := stringAttribute(t, record, "Bar"); got != "baz" {
		t.Errorf("Unexpected Bar after update: got %q, want %q", got, "baz")
	}

	// List
	filterExpression := "Id = :id"
	records, err := table.ListRecordsWithFilters(&DynamoDBFilter{
		FilterExpression: &filterExpression,
		AttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: item.Id},
		},
	})
	if err != nil {
		t.Fatalf("Failed listing dynamodb records: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("Unexpected number of listed records: got %v, want 1", len(records))
	}
	if got := stringAttribute(t, records[0], "Bar"); got != "baz" {
		t.Errorf("Unexpected Bar in listed record: got %q, want %q", got, "baz")
	}

	// Delete
	if err := table.DeleteRecord(item); err != nil {
		t.Fatalf("Failed deleting dynamodb record: %v", err)
	}
	record, err = table.GetRecord(item)
	if err != nil {
		t.Fatalf("Failed retrieving dynamodb record after delete: %v", err)
	}
	if record != nil {
		t.Errorf("Record %v still exists after deleting it", item.Id)
	}
}
