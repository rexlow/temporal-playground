package temporal

import (
	"log"

	"go.temporal.io/sdk/client"
)

type ClientManager struct {
	client client.Client
}

func NewClientManager(options client.Options) *ClientManager {

	temporalClient, err := client.Dial(options)
	if err != nil {
		log.Fatalf("Failed to create Temporal client for namespace '%s': %v", options.Namespace, err)
	}
	log.Printf("Connected to Temporal namespace: %s", options.Namespace)

	return &ClientManager{
		client: temporalClient,
	}
}

func NewAdminClientManager(options client.Options) *ClientManager {
	temporalClient, err := client.Dial(options)
	if err != nil {
		log.Fatalf("Failed to create Temporal admin client: %v", err)
	}

	return &ClientManager{
		client: temporalClient,
	}
}

func (cm *ClientManager) GetClient() client.Client {
	return cm.client
}

func (cm *ClientManager) NewScheduleClient() client.ScheduleClient {
	return cm.client.ScheduleClient()
}

func (cm *ClientManager) Close() {
	if cm.client != nil {
		cm.client.Close()
	}
}
