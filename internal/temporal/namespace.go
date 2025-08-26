package temporal

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/types/known/durationpb"
)

type NamespaceManager struct {
	adminClient *ClientManager
}

func NewNamespaceManager(options client.Options) *NamespaceManager {
	return &NamespaceManager{
		adminClient: NewAdminClientManager(options),
	}
}

func (nm *NamespaceManager) RegisterNamespace(namespaceName string, description string, retentionDays int) error {
	defer nm.adminClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if retentionDays <= 0 {
		retentionDays = 7 // Default to 7 days
	}

	retention := time.Duration(retentionDays) * 24 * time.Hour

	request := &workflowservice.RegisterNamespaceRequest{
		Namespace:                        namespaceName,
		Description:                      description,
		WorkflowExecutionRetentionPeriod: durationpb.New(retention),
	}

	if _, err := nm.adminClient.GetClient().WorkflowService().RegisterNamespace(ctx, request); err != nil {
		return fmt.Errorf("failed to register namespace '%s': %w", namespaceName, err)
	}

	log.Printf("Successfully registered namespace: %s", namespaceName)
	return nil
}

func (nm *NamespaceManager) ListNamespaces() ([]string, error) {
	defer nm.adminClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	request := &workflowservice.ListNamespacesRequest{}
	response, err := nm.adminClient.GetClient().WorkflowService().ListNamespaces(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	var namespaces []string
	for _, ns := range response.Namespaces {
		namespaces = append(namespaces, ns.NamespaceInfo.Name)
	}

	return namespaces, nil
}
