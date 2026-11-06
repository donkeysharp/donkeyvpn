package service_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/donkeysharp/donkeyvpn/internal/models"
	"github.com/donkeysharp/donkeyvpn/internal/service"
	"github.com/donkeysharp/donkeyvpn/internal/testutil"
)

func seed(t *testing.T, table *testutil.FakeTable, instances ...*models.VPNInstance) {
	t.Helper()
	for _, instance := range instances {
		if _, err := table.CreateRecord(instance); err != nil {
			t.Fatal(err)
		}
	}
}

func TestVPNServiceCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		table := testutil.NewFakeTable()
		seed(t, table, models.NewVPNInstance("vpn001", "vpn.example.com", "51820", "success", "i-1", "42"))
		asg := &testutil.FakeScalingGroup{DesiredCapacity: 1, MaxSize: 5}
		svc := service.NewVPNService(asg, table, nil)

		created, err := svc.Create(-100123)
		if err != nil || !created {
			t.Fatalf("expected created=true err=nil, got created=%v err=%v", created, err)
		}
		if !slices.Equal(asg.CapacityUpdates, []int32{2}) {
			t.Errorf("expected desired capacity to be raised to 2, got updates %v", asg.CapacityUpdates)
		}
		got, err := svc.Get("vpn002")
		if err != nil {
			t.Fatal(err)
		}
		want := *models.NewVPNInstance("vpn002", "pending", "pending", "pending", "pending", "-100123")
		if *got != want {
			t.Errorf("expected new record %v, got %v", want, got)
		}
	})

	t.Run("pending instance exists", func(t *testing.T) {
		table := testutil.NewFakeTable()
		seed(t, table, models.NewVPNInstance("vpn001", "pending", "pending", service.STATUS_PENDING, "pending", "42"))
		asg := &testutil.FakeScalingGroup{DesiredCapacity: 1, MaxSize: 5}
		svc := service.NewVPNService(asg, table, nil)

		created, err := svc.Create(42)
		if err != service.ErrVPNInstanceCreating || created {
			t.Fatalf("expected ErrVPNInstanceCreating, got created=%v err=%v", created, err)
		}
		if len(asg.CapacityUpdates) != 0 || table.Len() != 1 {
			t.Errorf("expected no side effects, got capacity updates %v and %d records", asg.CapacityUpdates, table.Len())
		}
	})

	t.Run("max capacity", func(t *testing.T) {
		table := testutil.NewFakeTable()
		asg := &testutil.FakeScalingGroup{DesiredCapacity: 2, MaxSize: 2}
		svc := service.NewVPNService(asg, table, nil)

		created, err := svc.Create(42)
		if err != service.ErrMaxCapacity || created {
			t.Fatalf("expected ErrMaxCapacity, got created=%v err=%v", created, err)
		}
		if len(asg.CapacityUpdates) != 0 || table.Len() != 0 {
			t.Errorf("expected no side effects, got capacity updates %v and %d records", asg.CapacityUpdates, table.Len())
		}
	})
}

func TestVPNServiceNextId(t *testing.T) {
	cases := []struct {
		name     string
		existing []string
		want     string
		wantErr  error
	}{
		{name: "empty table", want: "vpn001"},
		{name: "fills the first gap", existing: []string{"vpn001", "vpn003"}, want: "vpn002"},
		{name: "after the last", existing: []string{"vpn001", "vpn002"}, want: "vpn003"},
		{name: "only the first is free", existing: []string{"vpn002", "vpn003", "vpn004", "vpn005"}, want: "vpn001"},
		{name: "all ids taken", existing: []string{"vpn001", "vpn002", "vpn003", "vpn004", "vpn005"}, wantErr: service.ErrMaxCapacity},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			table := testutil.NewFakeTable()
			for _, id := range tc.existing {
				seed(t, table, models.NewVPNInstance(id, "", "", "success", "", ""))
			}
			got, err := service.NewVPNService(nil, table, nil).NextId()
			if got != tc.want || err != tc.wantErr {
				t.Errorf("expected %q %v, got %q %v", tc.want, tc.wantErr, got, err)
			}
		})
	}
}

func TestVPNServiceListOlderThan(t *testing.T) {
	now := time.Now()
	table := testutil.NewFakeTable()
	seed(t, table,
		models.NewVPNInstance("vpn001", "", "", "success", "i-old", "42"),
		models.NewVPNInstance("vpn002", "", "", "success", "i-new", "42"),
		models.NewVPNInstance("vpn003", "", "", "success", "i-stopped", "42"),
		// not described at all: the describer would fail on this id
		models.NewVPNInstance("vpn004", "", "", service.STATUS_PENDING, "i-unknown", "42"),
	)
	ec2 := &testutil.FakeInstanceDescriber{Instances: map[string]ec2types.Instance{
		"i-old":     testutil.NewEC2Instance("i-old", ec2types.InstanceStateNameRunning, now.Add(-2*time.Hour)),
		"i-new":     testutil.NewEC2Instance("i-new", ec2types.InstanceStateNameRunning, now.Add(-30*time.Minute)),
		"i-stopped": testutil.NewEC2Instance("i-stopped", ec2types.InstanceStateNameStopped, now.Add(-3*time.Hour)),
	}}

	got, err := service.NewVPNService(nil, table, ec2).ListOlderThan(time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Id != "vpn001" {
		t.Errorf("expected only vpn001, got %v", got)
	}
}

func TestVPNServiceListOlderThanDescribeError(t *testing.T) {
	table := testutil.NewFakeTable()
	seed(t, table, models.NewVPNInstance("vpn001", "", "", "success", "i-1", "42"))
	ec2 := &testutil.FakeInstanceDescriber{Err: fmt.Errorf("throttled")}

	if _, err := service.NewVPNService(nil, table, ec2).ListOlderThan(time.Hour); err != ec2.Err {
		t.Errorf("expected the describe error, got %v", err)
	}
}
