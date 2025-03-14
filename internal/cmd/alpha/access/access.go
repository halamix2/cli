package access

import (
	"fmt"

	"github.com/kyma-project/cli.v3/internal/clierror"
	"github.com/kyma-project/cli.v3/internal/cmdcommon"
	"github.com/kyma-project/cli.v3/internal/kube"
	"github.com/kyma-project/cli.v3/internal/kube/resources"
	"github.com/kyma-project/cli.v3/internal/kubeconfig"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
)

type accessConfig struct {
	*cmdcommon.KymaConfig

	name        string
	clusterrole string
	output      string
	namespace   string
	time        string
	permanent   bool
}

func NewAccessCMD(kymaConfig *cmdcommon.KymaConfig) *cobra.Command {
	cfg := accessConfig{
		KymaConfig: kymaConfig,
	}

	cmd := &cobra.Command{
		Use:   "access [flags]",
		Short: "Produces a kubeconfig with a Service Account-based token and certificate",
		Long:  "Use this command to produce a kubeconfig with a Service Account-based token and certificate that is valid for a specified time or indefinitely.",
		Run: func(_ *cobra.Command, _ []string) {
			clierror.Check(runAccess(&cfg))
		},
	}

	cmd.Flags().StringVar(&cfg.name, "name", "", "Name of the Service Account to be created")
	cmd.Flags().StringVar(&cfg.clusterrole, "clusterrole", "", "Name of the cluster role to bind the Service Account to")
	cmd.Flags().StringVar(&cfg.output, "output", "", "Path to the kubeconfig file output. If not provided, the kubeconfig will be printed")
	cmd.Flags().StringVar(&cfg.namespace, "namespace", "default", "Namespace in which the resource is created")
	cmd.Flags().StringVar(&cfg.time, "time", "1h", "Determines how long the token should be valid, by default 1h (use h for hours and d for days)")
	cmd.Flags().BoolVar(&cfg.permanent, "permanent", false, "Determines if the token is valid indefinitely")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("clusterrole")

	return cmd
}

func runAccess(cfg *accessConfig) clierror.Error {
	// Create objects
	clierr := createObjects(cfg)
	if clierr != nil {
		return clierror.WrapE(clierr, clierror.New("failed to create objects"))
	}

	// Fill kubeconfig
	generatedKubeconfig, clierr := kubeconfig.Prepare(cfg.Ctx, cfg.KubeClient, cfg.name, cfg.namespace, cfg.time, cfg.output, cfg.permanent)
	if clierr != nil {
		return clierr
	}

	// Print or write to file
	if cfg.output != "" {
		err := kube.SaveConfig(generatedKubeconfig, cfg.output)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to save kubeconfig"))
		}
	} else {
		message, err := clientcmd.Write(*generatedKubeconfig)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to print kubeconfig"))
		}
		fmt.Println(string(message))
	}
	return nil
}

func createObjects(cfg *accessConfig) clierror.Error {
	// Create Service Account
	err := resources.CreateServiceAccount(cfg.Ctx, cfg.KubeClient, cfg.name, cfg.namespace)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create Service Account"))
	}
	// Create Role Binding for the Service Account
	err = resources.CreateClusterRoleBinding(cfg.Ctx, cfg.KubeClient, cfg.name, cfg.namespace, cfg.clusterrole)
	if err != nil {
		return clierror.Wrap(err, clierror.New("failed to create Cluster Role Binding"))
	}
	// Create a service-account-token type secret
	if cfg.permanent {
		err = resources.CreateServiceAccountToken(cfg.Ctx, cfg.KubeClient, cfg.name, cfg.namespace)
		if err != nil {
			return clierror.Wrap(err, clierror.New("failed to create secret"))
		}
	}
	return nil
}
