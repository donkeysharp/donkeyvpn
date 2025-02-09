package service

import (
	"fmt"

	"github.com/donkeysharp/donkeyvpn/internal/aws"
	"github.com/donkeysharp/donkeyvpn/internal/models"
	"github.com/labstack/gommon/log"
)

const STATUS_PENDING = "pending"
const STATUS_READY = "ready"
const MAX_INSTANCEs = 5
const UNDEFINED = "undefined"

func NewVPNService(asg *aws.AutoscalingGroup, table *aws.DynamoDB) *VPNService {
	return &VPNService{
		asg:   asg,
		table: table,
	}
}

type VPNService struct {
	asg   *aws.AutoscalingGroup
	table *aws.DynamoDB
}

var ErrMaxCapacity = fmt.Errorf("ASG reached its maximum of instances")
var ErrVPNInstanceNotFound = fmt.Errorf("VPN Instance does not exist")

func allPossibleIds() []string {
	var possibleIds []string = make([]string, 0)
	for i := 1; i <= MAX_INSTANCEs; i++ {
		possibleIds = append(possibleIds, fmt.Sprintf("vpn%03d", i))
	}
	return possibleIds
}

func (s *VPNService) NextId() (string, error) {
	possibleIds := allPossibleIds()
	instances, err := s.ListMap()
	if err != nil {
		return "", err
	}

	for _, id := range possibleIds {
		if _, exists := instances[id]; !exists {
			return id, nil
		}
	}
	return "", ErrMaxCapacity
}

func (s *VPNService) UpdateMapper(item models.ModelMapper) error {
	model := item.ToModel().(models.VPNInstance)
	return s.Update(model)
}

func (s *VPNService) Update(item models.VPNInstance) error {
	instance, err := s.Get(item.Id)
	if err != nil {
		return err
	}

	if instance == nil {
		return ErrVPNInstanceNotFound
	}

	instance.Hostname = item.Hostname
	instance.Port = item.Port
	instance.Status = item.Status
	instance.InstanceId = item.InstanceId

	return s.table.UpdateRecord(instance)
}

func (s *VPNService) Create() (bool, error) {
	asg, err := s.asg.GetInfo()
	if err != nil {
		log.Error("Error while getting ASG information")
		return false, err
	}
	log.Infof("ASG desiredCapacity: %d, maxSize: %d", *asg.DesiredCapacity, *asg.MaxSize)
	if *asg.DesiredCapacity == *asg.MaxSize {
		log.Warnf(ErrMaxCapacity.Error())
		return false, ErrMaxCapacity
	}

	desiredCapacity := *asg.DesiredCapacity + 1
	err = s.asg.UpdateCapacity(int32(desiredCapacity))
	if err != nil {
		log.Errorf("Error while updating capacity of ASG to %d", desiredCapacity)
		return false, err
	}

	nextId, err := s.NextId()
	log.Infof("Next possible id: %s", nextId)
	if err != nil {
		return false, err
	}

	instance := models.NewVPNInstance(nextId, UNDEFINED, UNDEFINED, STATUS_PENDING, UNDEFINED)
	log.Infof("Creating a new record for id: %s", nextId)
	result, err := s.table.CreateRecord(instance)
	if err != nil {
		return false, err
	}
	log.Infof("Record creation result: %v", result)

	return result, nil
}

func (s *VPNService) ListArray() ([]models.VPNInstance, error) {
	itemsRaw, err := s.table.ListRecords()
	if err != nil {
		return nil, err
	}

	instances, err := models.DynamoItemsToVPNInstances(itemsRaw)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (s *VPNService) Get(id string) (*models.VPNInstance, error) {
	item, err := s.table.GetRecord(models.VPNInstance{Id: id})

	if err != nil {
		log.Errorf("Failed to get VPN instance with id %v: %v", id, err.Error())
		return nil, err
	}
	if item == nil {
		log.Warnf("VPN Instance with id %v not found", id)
		return nil, ErrVPNInstanceNotFound
	}

	instance, err := models.DynamoItemToVPNInstance(item)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (s *VPNService) ListMap() (map[string]models.VPNInstance, error) {
	itemsRaw, err := s.table.ListRecords()
	if err != nil {
		return nil, err
	}

	instances, err := models.DynamoItemsToVPNInstancesMap(itemsRaw)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (s *VPNService) Delete(vpnId string) (bool, error) {
	instance, err := s.Get(vpnId)
	if err != nil {
		return false, err
	}
	// TODO: delete in ASG as well
	err = s.table.DeleteRecord(instance)
	if err != nil {
		log.Errorf("Failed to delete vpn instance %v. Error: %v", instance, err.Error())
		return false, err
	}
	log.Infof("Instance with id %v deleted successfully", vpnId)
	return true, nil
}
