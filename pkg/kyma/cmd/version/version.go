package version

import (
	"fmt"
	"strings"
	"time"

	"github.com/kyma-project/cli/internal/kube"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"
)

//Version contains the cli binary version injected by the build system
var Version string

type command struct {
	opts *Options
}

//NewVersionCmd creates a new version command
func NewCmd(o *Options) *cobra.Command {
	c := &command{
		opts: o,
	}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Version of the kyma CLI and connected Kyma cluster",
		Long: `Prints the version of kyma CLI itself and the version of the kyma cluster connected by current KUBECONFIG
`,
		RunE: func(_ *cobra.Command, _ []string) error { return c.Run() },
	}
	cmd.Flags().BoolVarP(&o.Client, "client", "c", false, "Client version only (no server required)")

	return cmd
}

//Run runs the command
func (c command) Run() error {
	version := Version
	if version == "" {
		version = "N/A"
	}
	fmt.Printf("Kyma CLI version: %s\n", version)

	if !c.opts.Client {
		k8s, err := kube.NewFromConfigWithTimeout("", c.opts.KubeconfigPath, 2*time.Second)
		if err != nil {
			return errors.Wrap(err, "Could not initialize the Kubernetes client. PLease make sure that you have a valid kubeconfig.")
		}

		version, err := KymaVersion(c.opts.Verbose, k8s)
		if err != nil {
			fmt.Printf("Unable to get Kyma cluster version due to error: %s. Please check if your cluster is available and has Kyma installed\r\n", err.Error())
			return nil
		}
		fmt.Printf("Kyma cluster version: %s\n", version)
	}

	return nil
}

//KymaVersion determines the version of kyma installed in the cluster sccessible via the provided kubernetes client
func KymaVersion(verbose bool, k8s kube.KymaKube) (string, error) {
	//kymaVersion, err := kubectl.RunCmdWithTimeout(2*time.Second, verbose, "-n", "kyma-installer", "get", "pod", "-l", "name=kyma-installer", "-o", "jsonpath='{.items[*].spec.containers[0].image}'")
	pods, err := k8s.CoreV1().Pods("kyma-installer").List(metav1.ListOptions{LabelSelector: "name=kyma-installer"})
	if err != nil {
		return "", err
	}

	if len(pods.Items) == 0 {
		return "N/A", nil
	}

	kymaVersion := pods.Items[0].Spec.Containers[0].Image
	return strings.Split(kymaVersion, ":")[1], nil
}
