// Package testutil holds hand-written in-memory fakes for the AWS and Telegram
// clients, so services, processors and handlers can be unit tested without
// network access. Each fake also has an Err field: when it is set, every
// method returns it.
package testutil

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/donkeysharp/donkeyvpn/internal/aws"
)

// FakeTable is an in-memory DynamoDB table with the same methods as
// *aws.DynamoDB. Items are stored as attribute maps keyed by their primary
// key, so data goes through the real marshalling code in internal/models.
//
// UpdateRecord supports "SET #name = :value, ..." expressions and
// ListRecordsWithFilters supports "#name = :value [AND ...]" filters, which is
// what internal/models generates. Anything else returns an error.
type FakeTable struct {
	mu    sync.Mutex
	items map[string]map[string]types.AttributeValue
	Err   error
}

func NewFakeTable() *FakeTable {
	return &FakeTable{items: make(map[string]map[string]types.AttributeValue)}
}

func keyOf(pk map[string]types.AttributeValue) (string, error) {
	if len(pk) == 0 {
		return "", fmt.Errorf("fake table: empty primary key")
	}
	parts := make([]string, 0, len(pk))
	for name, value := range pk {
		switch v := value.(type) {
		case *types.AttributeValueMemberS:
			parts = append(parts, name+"="+v.Value)
		case *types.AttributeValueMemberN:
			parts = append(parts, name+"="+v.Value)
		default:
			return "", fmt.Errorf("fake table: unsupported key type %T for %s", value, name)
		}
	}
	sort.Strings(parts)
	return strings.Join(parts, ","), nil
}

func copyItem(item map[string]types.AttributeValue) map[string]types.AttributeValue {
	out := make(map[string]types.AttributeValue, len(item))
	for k, v := range item {
		out[k] = v
	}
	return out
}

func (t *FakeTable) CreateRecord(item aws.DynamoDBItem) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Err != nil {
		return false, t.Err
	}
	key, err := keyOf(item.PrimaryKey())
	if err != nil {
		return false, err
	}
	t.items[key] = copyItem(item.ToItem())
	return true, nil
}

func (t *FakeTable) GetRecord(item aws.DynamoDBItem) (map[string]types.AttributeValue, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Err != nil {
		return nil, t.Err
	}
	key, err := keyOf(item.PrimaryKey())
	if err != nil {
		return nil, err
	}
	stored, ok := t.items[key]
	if !ok {
		// *aws.DynamoDB returns nil, nil for a missing item
		return nil, nil
	}
	return copyItem(stored), nil
}

// UpdateRecord sets the attributes named in the update expression. Like
// DynamoDB UpdateItem, it creates the item when it does not exist.
func (t *FakeTable) UpdateRecord(item aws.UpdateItem) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Err != nil {
		return t.Err
	}
	expr, err := item.UpdateExpression()
	if err != nil {
		return err
	}
	pk := item.PrimaryKey()
	key, err := keyOf(pk)
	if err != nil {
		return err
	}

	update := strings.TrimSpace(*expr.Update())
	if !strings.HasPrefix(update, "SET ") {
		return fmt.Errorf("fake table: unsupported update expression %q", update)
	}
	stored, ok := t.items[key]
	if !ok {
		stored = copyItem(pk)
	} else {
		stored = copyItem(stored)
	}
	for _, assignment := range strings.Split(strings.TrimPrefix(update, "SET "), ",") {
		name, value, err := resolve(assignment, expr.Names(), expr.Values())
		if err != nil {
			return err
		}
		stored[name] = value
	}
	t.items[key] = stored
	return nil
}

func (t *FakeTable) ListRecords() ([]map[string]types.AttributeValue, error) {
	return t.ListRecordsWithFilters(nil)
}

// ListRecordsWithFilters returns matching items ordered by primary key, so
// results are deterministic (a real Scan has no defined order).
func (t *FakeTable) ListRecordsWithFilters(filter *aws.DynamoDBFilter) ([]map[string]types.AttributeValue, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Err != nil {
		return nil, t.Err
	}

	type condition struct {
		name  string
		value types.AttributeValue
	}
	var conditions []condition
	if filter != nil && filter.FilterExpression != nil {
		for _, part := range strings.Split(*filter.FilterExpression, " AND ") {
			name, value, err := resolve(part, filter.AttributeNames, filter.AttributeValues)
			if err != nil {
				return nil, err
			}
			conditions = append(conditions, condition{name, value})
		}
	}

	keys := make([]string, 0, len(t.items))
	for key := range t.items {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]map[string]types.AttributeValue, 0, len(keys))
	for _, key := range keys {
		item := t.items[key]
		matches := true
		for _, c := range conditions {
			if !reflect.DeepEqual(item[c.name], c.value) {
				matches = false
				break
			}
		}
		if matches {
			result = append(result, copyItem(item))
		}
	}
	return result, nil
}

func (t *FakeTable) DeleteRecord(item aws.DynamoDBItem) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.Err != nil {
		return t.Err
	}
	key, err := keyOf(item.PrimaryKey())
	if err != nil {
		return err
	}
	delete(t.items, key)
	return nil
}

// Len returns the number of stored items.
func (t *FakeTable) Len() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.items)
}

// resolve turns "#name = :value" into the attribute name and value it refers
// to. A name without a "#" placeholder is used as is.
func resolve(expr string, names map[string]string, values map[string]types.AttributeValue) (string, types.AttributeValue, error) {
	left, right, ok := strings.Cut(expr, "=")
	if !ok {
		return "", nil, fmt.Errorf("fake table: unsupported expression %q", expr)
	}
	left, right = strings.TrimSpace(left), strings.TrimSpace(right)
	name := left
	if strings.HasPrefix(left, "#") {
		if name, ok = names[left]; !ok {
			return "", nil, fmt.Errorf("fake table: unknown attribute name %q", left)
		}
	}
	value, ok := values[right]
	if !ok {
		return "", nil, fmt.Errorf("fake table: unknown attribute value %q", right)
	}
	return name, value, nil
}
