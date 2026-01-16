package exporter

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
)

// createOTELResource creates an OTEL resource from configuration attributes.
func createOTELResource(resourceAttrs map[string]string) (*resource.Resource, error) {
	attrs := make([]attribute.KeyValue, 0, len(resourceAttrs))
	for k, v := range resourceAttrs {
		attrs = append(attrs, attribute.String(k, v))
	}

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(attrs...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	return res, nil
}
