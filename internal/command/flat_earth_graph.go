package command

import (
	"github.com/hashicorp/terraform/internal/backend"
	"github.com/hashicorp/terraform/internal/configs"
	"github.com/hashicorp/terraform/internal/plans"
	"github.com/hashicorp/terraform/internal/plans/planfile"
	"github.com/hashicorp/terraform/internal/tfdiags"
)

type FlatEarthGraphCommand struct {
	Meta
}

// this is a command only in name, we don't want to waste our time writing for a terminal's UI
func (c *FlatEarthGraphCommand) Run(configPath string) (map[string]*configs.Resource, tfdiags.Diagnostics) {
	var diags tfdiags.Diagnostics
	var err error
	var planPath string
	var planFile *planfile.Reader

	if planPath != "" {
		planFile, err = c.PlanFile(planPath)
		if err != nil {
			diags.Append(err)
			return nil, diags
		}
	}

	backendConfig, backendDiags := c.loadBackendConfig(configPath)
	diags = diags.Append(backendDiags)
	if diags.HasErrors() {
		return nil, diags
	}

	b, backendDiags := c.Backend(&BackendOpts{
		Config: backendConfig,
	})
	diags = diags.Append(backendDiags)
	if diags.HasErrors() {
		return nil, diags
	}

	local, ok := b.(backend.Local)
	if !ok {
		if diags.HasErrors() {
			return nil, diags
		}
	}

	c.ignoreRemoteVersionConflict(b)

	opReq := c.Operation(b)
	opReq.ConfigDir = configPath
	opReq.ConfigLoader, err = c.initConfigLoader()
	opReq.PlanFile = planFile
	opReq.AllowUnsetVariables = true
	if err != nil {
		diags = diags.Append(err)
		return nil, diags
	}

	// Get the context
	lr, _, ctxDiags := local.LocalRun(opReq)
	diags = diags.Append(ctxDiags)
	if ctxDiags.HasErrors() {
		return nil, diags
	}

	// TODO: we should probably consider extending this to other plan modes, see graph command
	graph := lr.Core.PlanFlatEarthGraph(lr.Config, lr.InputState, plans.NormalMode)
	return graph, diags
}
