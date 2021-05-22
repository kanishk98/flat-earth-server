package main

import (
	"os"
	"os/signal"

	"github.com/hashicorp/go-plugin"
	svchost "github.com/hashicorp/terraform-svchost"
	"github.com/hashicorp/terraform-svchost/auth"
	"github.com/hashicorp/terraform-svchost/disco"
	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/command"
	"github.com/hashicorp/terraform/command/cliconfig"
	"github.com/hashicorp/terraform/internal/getproviders"
	pluginDiscovery "github.com/hashicorp/terraform/plugin/discovery"
	"github.com/hashicorp/terraform/terraform"
)

func getCommands(
	originalWorkingDir string,
	config *cliconfig.Config,
	services *disco.Disco,
	providerSrc getproviders.Source,
	providerDevOverrides map[addrs.Provider]getproviders.PackageLocalDir,
	unmanagedProviders map[addrs.Provider]*plugin.ReattachConfig,
) map[string]FlatEarthCommand {

	for userHost, hostConfig := range config.Hosts {
		host, err := svchost.ForComparison(userHost)
		if err != nil {
			// We expect the config was already validated by the time we get
			// here, so we'll just ignore invalid hostnames.
			continue
		}
		services.ForceHostServices(host, hostConfig.Services)
	}

	configDir, err := cliconfig.ConfigDir()
	if err != nil {
		configDir = "" // No config dir available (e.g. looking up a home directory failed)
	}

	dataDir := os.Getenv("TF_DATA_DIR")

	meta := command.Meta{
		OriginalWorkingDir: originalWorkingDir,

		GlobalPluginDirs: globalPluginDirs(),

		Services: services,

		RunningInAutomation: false,
		CLIConfigDir:        configDir,
		PluginCacheDir:      config.PluginCacheDir,
		OverrideDataDir:     dataDir,

		ShutdownCh: makeShutdownCh(),

		ProviderSource:       providerSrc,
		ProviderDevOverrides: providerDevOverrides,
		UnmanagedProviders:   unmanagedProviders,
	}

	// The command list is included in the terraform -help
	// output, which is in turn included in the docs at
	// website/docs/cli/commands/index.html.markdown; if you
	// add, remove or reclassify commands then consider updating
	// that to match.

	commands := map[string]FlatEarthCommand{
		"flat-earth": &command.FlatEarthGraphCommand{
			Meta: meta,
		},
	}
	return commands
}

// makeShutdownCh creates an interrupt listener and returns a channel.
// A message will be sent on the channel for every interrupt received.
func makeShutdownCh() <-chan struct{} {
	resultCh := make(chan struct{})

	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, ignoreSignals...)
	signal.Notify(signalCh, forwardSignals...)
	go func() {
		for {
			<-signalCh
			resultCh <- struct{}{}
		}
	}()

	return resultCh
}

func credentialsSource(config *cliconfig.Config) (auth.CredentialsSource, error) {
	helperPlugins := pluginDiscovery.FindPlugins("credentials", globalPluginDirs())
	return config.CredentialsSource(helperPlugins)
}

type FlatEarthCommand interface {
	Run(configPath string) (*terraform.Graph, error)
}
