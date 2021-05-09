package command

import (
	"github.com/hashicorp/terraform/tfdiags"

	"github.com/hashicorp/terraform/backend"
	"github.com/hashicorp/terraform/terraform"
)

// FlatEarthGraphCommand is a Command implementation that takes a Terraform
// configuration and outputs the dependency tree in graphical form.
type FlatEarthGraphCommand struct {
	Meta
}

func (c *FlatEarthGraphCommand) Run(configPath string) ([]terraform.GraphTransformer, error) {

	var diags tfdiags.Diagnostics

	backendConfig, backendDiags := c.loadBackendConfig(configPath)
	diags = diags.Append(backendDiags)
	if diags.HasErrors() {
		c.showDiagnostics(diags)
		return nil, diags.Err()
	}

	// Load the backend
	b, backendDiags := c.Backend(&BackendOpts{
		Config: backendConfig,
	})
	diags = diags.Append(backendDiags)
	if backendDiags.HasErrors() {
		c.showDiagnostics(diags)
		return nil, diags.Err()
	}

	// We require a local backend
	local, ok := b.(backend.Local)
	if !ok {
		c.showDiagnostics(diags) // in case of any warnings in here
		c.Ui.Error(ErrUnsupportedLocalOp)
		return nil, diags.Err()
	}

	// This is a read-only command
	c.ignoreRemoteBackendVersionConflict(b)

	// Build the operation
	err := error(nil)
	opReq := c.Operation(b)
	opReq.ConfigDir = configPath
	opReq.ConfigLoader, err = c.initConfigLoader()
	opReq.AllowUnsetVariables = true
	if err != nil {
		diags = diags.Append(err)
		c.showDiagnostics(diags)
		return nil, diags.Err()
	}

	// Get the context
	ctx, _, ctxDiags := local.Context(opReq)
	diags = diags.Append(ctxDiags)
	if ctxDiags.HasErrors() {
		c.showDiagnostics(diags)
		return nil, ctxDiags.Err()
	}

	// Skip validation during graph generation - we want to see the graph even if
	// it is invalid for some reason.
	return ctx.FlatEarthGraph(), error(nil)
}
