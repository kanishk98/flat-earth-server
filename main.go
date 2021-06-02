package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform-svchost/disco"
	"github.com/hashicorp/terraform/addrs"
	"github.com/hashicorp/terraform/command/cliconfig"
	"github.com/hashicorp/terraform/command/format"
	"github.com/hashicorp/terraform/httpclient"
	"github.com/hashicorp/terraform/internal/logging"
	"github.com/hashicorp/terraform/version"
	"github.com/mitchellh/colorstring"

	backendInit "github.com/hashicorp/terraform/backend/init"
)

const (
	// The parent process will create a file to collect crash logs
	envTmpLogPath = "TF_TEMP_LOG_PATH"
)

var commands map[string]FlatEarthCommand
var args []string

func main() {
	os.Exit(wrappedMain())
}

func wrappedMain() int {
	commands, args := prepCommands()
	// kanishk98: This is a bit of an anti-pattern. Fix this.
	if args == nil {
		return 1
	}
	startServer(commands, args)
	return 0
}

func prepCommands() (map[string]FlatEarthCommand, []string) {
	var err error

	tmpLogPath := os.Getenv(envTmpLogPath)
	if tmpLogPath != "" {
		f, err := os.OpenFile(tmpLogPath, os.O_RDWR|os.O_APPEND, 0666)
		if err == nil {
			defer f.Close()

			log.Printf("[DEBUG] Adding temp file log sink: %s", f.Name())
			logging.RegisterSink(f)
		} else {
			log.Printf("[ERROR] Could not open temp log file: %v", err)
		}
	}

	// NOTE: We're intentionally calling LoadConfig _before_ handling a possible
	// -chdir=... option on the command line, so that a possible relative
	// path in the TERRAFORM_CONFIG_FILE environment variable (though probably
	// ill-advised) will be resolved relative to the true working directory,
	// not the overridden one.
	config, _ := cliconfig.LoadConfig()

	// Get any configured credentials from the config and initialize
	// a service discovery object. The slightly awkward predeclaration of
	// disco is required to allow us to pass untyped nil as the creds source
	// when creating the source fails. Otherwise we pass a typed nil which
	// breaks the nil checks in the disco object
	var services *disco.Disco
	credsSrc, err := credentialsSource(config)
	if err == nil {
		services = disco.NewWithCredentialsSource(credsSrc)
	} else {
		// Most commands don't actually need credentials, and most situations
		// that would get us here would already have been reported by the config
		// loading above, so we'll just log this one as an aid to debugging
		// in the unlikely event that it _does_ arise.
		log.Printf("[WARN] Cannot initialize remote host credentials manager: %s", err)
		// passing (untyped) nil as the creds source is okay because the disco
		// object checks that and just acts as though no credentials are present.
		services = disco.NewWithCredentialsSource(nil)
	}
	services.SetUserAgent(httpclient.TerraformUserAgent(version.String()))

	providerSrc, diags := providerSource(config.ProviderInstallation, services)
	if len(diags) > 0 {
		log.Fatal("There are some problems with the provider_installation configuration:")
		for _, diag := range diags {
			earlyColor := &colorstring.Colorize{
				Colors:  colorstring.DefaultColors,
				Disable: true, // Disable color to be conservative until we know better
				Reset:   true,
			}
			log.Fatal(format.Diagnostic(diag, nil, earlyColor, 78))
		}
		if diags.HasErrors() {
			log.Fatal("As a result of the above problems, Terraform's provider installer may not behave as intended.\n\n")
			// We continue to run anyway, because most commands don't do provider installation.
		}
	}
	providerDevOverrides := providerDevOverrides(config.ProviderInstallation)

	// The user can declare that certain providers are being managed on
	// Terraform's behalf using this environment variable. This is used
	// primarily by the SDK's acceptance testing framework.
	unmanagedProviders, err := parseReattachProviders(os.Getenv("TF_REATTACH_PROVIDERS"))
	if err != nil {
		log.Fatal(err.Error())
		return nil, nil
	}

	// Initialize the backends.
	backendInit.Init(services)

	args = os.Args[1:]
	originalWd := args[0]

	commands = getCommands(originalWd, config, services, providerSrc, providerDevOverrides, unmanagedProviders)
	return commands, args
}

// parse information on reattaching to unmanaged providers out of a
// JSON-encoded environment variable.
func parseReattachProviders(in string) (map[addrs.Provider]*plugin.ReattachConfig, error) {
	unmanagedProviders := map[addrs.Provider]*plugin.ReattachConfig{}
	if in != "" {
		type reattachConfig struct {
			Protocol string
			Addr     struct {
				Network string
				String  string
			}
			Pid  int
			Test bool
		}
		var m map[string]reattachConfig
		err := json.Unmarshal([]byte(in), &m)
		if err != nil {
			return unmanagedProviders, fmt.Errorf("invalid format for TF_REATTACH_PROVIDERS: %w", err)
		}
		for p, c := range m {
			a, diags := addrs.ParseProviderSourceString(p)
			if diags.HasErrors() {
				return unmanagedProviders, fmt.Errorf("error parsing %q as a provider address: %w", a, diags.Err())
			}
			var addr net.Addr
			switch c.Addr.Network {
			case "unix":
				addr, err = net.ResolveUnixAddr("unix", c.Addr.String)
				if err != nil {
					return unmanagedProviders, fmt.Errorf("invalid unix socket path %q for %q: %w", c.Addr.String, p, err)
				}
			case "tcp":
				addr, err = net.ResolveTCPAddr("tcp", c.Addr.String)
				if err != nil {
					return unmanagedProviders, fmt.Errorf("invalid TCP address %q for %q: %w", c.Addr.String, p, err)
				}
			default:
				return unmanagedProviders, fmt.Errorf("unknown address type %q for %q", c.Addr.Network, p)
			}
			unmanagedProviders[a] = &plugin.ReattachConfig{
				Protocol: plugin.Protocol(c.Protocol),
				Pid:      c.Pid,
				Test:     c.Test,
				Addr:     addr,
			}
		}
	}
	return unmanagedProviders, nil
}
