package command

import (
	"errors"

	"github.com/hashicorp/terraform/backend"
	"github.com/hashicorp/terraform/configs"
)

// FlatEarthGraphCommand is a Command implementation that takes a Terraform
// configuration and outputs the dependency tree in graphical form.
type FlatEarthGraphCommand struct {
	Meta
}

func (c *FlatEarthGraphCommand) Run(configPath string) (map[string]*configs.Resource, error) {

	backendConfig, _ := c.loadBackendConfig(configPath)

	// Load the backend
	b, _ := c.Backend(&BackendOpts{
		Config: backendConfig,
	})

	// We require a local backend
	local, ok := b.(backend.Local)
	if !ok {
		return nil, errors.New(ErrUnsupportedLocalOp)
	}

	// This is a read-only command
	c.ignoreRemoteBackendVersionConflict(b)

	// Build the operation
	opReq := c.Operation(b)
	opReq.ConfigDir = configPath
	opReq.ConfigLoader, _ = c.initConfigLoader()
	opReq.AllowUnsetVariables = true

	// Get the context
	ctx, _, _ := local.Context(opReq)

	// Skip validation during graph generation - we want to see the graph even if
	// it is invalid for some reason.
	return ctx.FlatEarthGraph(), error(nil)
}
