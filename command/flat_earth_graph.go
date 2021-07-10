package command

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/backend"
	"github.com/hashicorp/terraform/command/jsonprovider"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/terraform"
)

// FlatEarthGraphCommand is a Command implementation that takes a Terraform
// configuration and outputs the dependency tree in graphical form.
type FlatEarthGraphCommand struct {
	Meta
}

func (c* FlatEarthGraphCommand) GetContext(configPath string) (*terraform.Context, error) {
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
	err := error(nil)
	opReq.ConfigDir = configPath
	opReq.ConfigLoader, err = c.initConfigLoader()
	if err != nil {
		return nil, err
	}
	opReq.AllowUnsetVariables = true
	// Get the context
	ctx, _, diags := local.Context(opReq)
	if diags.HasErrors() {
		return nil, diags.Err()
	}
	return ctx, nil
}

func (c *FlatEarthGraphCommand) Run(configPath string) (map[string]*configs.Resource, error) {
	ctx, err := c.GetContext(configPath)
	if err != nil {
		return nil, err
	}
	return ctx.FlatEarthGraph(), error(nil)
}

func (c *FlatEarthGraphCommand) GetProviderSchema(configPath string) ([]byte, error) {
	ctx, err := c.GetContext(configPath)
	if err != nil {
		return nil, err
	}
	schemas := ctx.Schemas()
	jsonSchemas, err := jsonprovider.Marshal(schemas)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to marshal provider schemas to json: %s", err))
		return nil, err
	}
	return jsonSchemas, nil
}
