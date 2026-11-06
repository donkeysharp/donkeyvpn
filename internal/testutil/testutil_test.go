package testutil

import (
	"testing"

	"github.com/donkeysharp/donkeyvpn/internal/models"
)

func TestFakeTableRoundTripsVPNInstance(t *testing.T) {
	table := NewFakeTable()
	want := models.NewVPNInstance("vpn001", "pending", "pending", "pending", "pending", "-100123")

	if ok, err := table.CreateRecord(want); err != nil || !ok {
		t.Fatalf("CreateRecord: ok=%v err=%v", ok, err)
	}

	item, err := table.GetRecord(models.VPNInstance{Id: "vpn001"})
	if err != nil || item == nil {
		t.Fatalf("GetRecord: item=%v err=%v", item, err)
	}
	got, err := models.DynamoItemToVPNInstance(item)
	if err != nil {
		t.Fatal(err)
	}
	if *got != *want {
		t.Errorf("expected %v, got %v", want, got)
	}

	// UpdateRecord only touches the attributes in the update expression
	update := models.VPNInstance{Id: "vpn001", Hostname: "vpn.example.com", Port: "51820", Status: "success", InstanceId: "i-123"}
	if err := table.UpdateRecord(update); err != nil {
		t.Fatal(err)
	}
	item, _ = table.GetRecord(models.VPNInstance{Id: "vpn001"})
	got, _ = models.DynamoItemToVPNInstance(item)
	update.ChatId = "-100123"
	if *got != update {
		t.Errorf("expected %v after update, got %v", update, got)
	}

	if err := table.DeleteRecord(got); err != nil {
		t.Fatal(err)
	}
	if item, err := table.GetRecord(models.VPNInstance{Id: "vpn001"}); item != nil || err != nil {
		t.Errorf("expected a deleted record to be missing, got item=%v err=%v", item, err)
	}
}

func TestFakeTableFiltersByStatus(t *testing.T) {
	table := NewFakeTable()
	table.CreateRecord(models.NewVPNInstance("vpn002", "", "", "pending", "", ""))
	table.CreateRecord(models.NewVPNInstance("vpn001", "", "", "success", "", ""))
	table.CreateRecord(models.NewVPNInstance("vpn003", "", "", "pending", "", ""))

	items, err := table.ListRecordsWithFilters(models.FilterInstanceByStatus("pending"))
	if err != nil {
		t.Fatal(err)
	}
	instances, _ := models.DynamoItemsToVPNInstances(items)
	if len(instances) != 2 || instances[0].Id != "vpn002" || instances[1].Id != "vpn003" {
		t.Errorf("expected [vpn002 vpn003], got %v", instances)
	}

	all, _ := table.ListRecords()
	if len(all) != 3 {
		t.Errorf("expected 3 records, got %d", len(all))
	}
}
