package remove

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	scheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	chaosClient "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/typed/litmuschaos/v1alpha1"
	restclient "k8s.io/client-go/rest"
)

var (
	kubeconfig string
	config     *restclient.Config
	client     *kubernetes.Clientset
	clientSet  *chaosClient.LitmuschaosV1alpha1Client
)

func TestChaos(t *testing.T) {

	RegisterFailHandler(Fail)
	RunSpecs(t, "BDD test")
}

var _ = BeforeSuite(func() {

	var err error
	kubeconfig = os.Getenv("HOME") + "/.kube/config"
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)

	if err != nil {
		Expect(err).To(BeNil(), "failed to get config")
	}

	client, err = kubernetes.NewForConfig(config)

	if err != nil {
		Expect(err).To(BeNil(), "failed to get client")
	}

	clientSet, err = chaosClient.NewForConfig(config)

	if err != nil {
		Expect(err).To(BeNil(), "failed to get clientSet")
	}

	err = v1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		fmt.Println(err)
	}

	//Deleting crds
	By("Creating chaosengine crd")
	err = exec.Command("kubectl", "delete", "-f", "https://raw.githubusercontent.com/litmuschaos/chaos-operator/master/deploy/chaos_crds.yaml").Run()
	Expect(err).To(BeNil(), "failed to delete crds")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("crds deleted successfully")

	//Deleting rbacs
	By("Creating chaosengine rbac")
	err = exec.Command("kubectl", "delete", "-f", "https://raw.githubusercontent.com/litmuschaos/chaos-operator/master/deploy/rbac.yaml").Run()
	Expect(err).To(BeNil(), "failed to create rbac")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("rbac deleted sucessfully")

	//Deleting chaos-operator deployment
	By("Deleting chaos-operator deployment")
	err = exec.Command("kubectl", "delete", "deploy", "-n", "litmus", "--all").Run()
	Expect(err).To(BeNil(), "failed to delete deployments")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("chaos-operator deleted sucessfully")

	//Checking Whether Litmus is Deleted successfully
	operator, err := client.AppsV1().Deployments("litmus").Get("chaos-operator", metav1.GetOptions{})
	count := 0
	for operator.Status.AvailableReplicas != 0 {
		if count > 5 {
			fmt.Printf("Avaliable Count: %v", operator.Status.AvailableReplicas)
			operator, _ = client.AppsV1().Deployments("litmus").Get("chaos-operator", metav1.GetOptions{})
			time.Sleep(2 * time.Second)
			count++
		} else {
			Fail("Litmus Deletion Failed Time Out")
		}
	}

	fmt.Println("Litmus is deleted successfully")

	//Removing Application
	By("Deleting Application,Service and PVC")
	err = exec.Command("kubectl", "delete", "-f", "percona.yaml").Run()
	Expect(err).To(BeNil(), "failed to delete application and its components")
	if err != nil {
		fmt.Println(err)
	}

	//Removing Other OpenEBS Componentes

	//Removing Blockdevices
	By("Deleting Blockdevices")
	err = exec.Command("kubectl", "delete", "-f", "bd", "-n", "openebs", "--all").Run()
	Expect(err).To(BeNil(), "failed to delete the blockdevices")
	if err != nil {
		fmt.Println(err)
	}

	//Removing BlockDeviceClaim
	By("Deleting BDC")
	err = exec.Command("kubectl", "delete", "-f", "bdc", "-n", "openebs", "--all").Run()
	Expect(err).To(BeNil(), "failed to delete the bdc")
	if err != nil {
		fmt.Println(err)
	}

	//Removing CSP
	By("Deleting CSP")
	err = exec.Command("kubectl", "delete", "-f", "csp", "--all").Run()
	Expect(err).To(BeNil(), "failed to delete csp")
	if err != nil {
		fmt.Println(err)
	}

	//Removing SPC
	By("Deleting SPC")
	err = exec.Command("kubectl", "delete", "-f", "spc", "--all").Run()
	Expect(err).To(BeNil(), "failed to delete spc")
	if err != nil {
		fmt.Println(err)
	}

	//Removing the deployment components
	By("Deleting deployment components")
	err = exec.Command("kubectl", "delete", "deploy", "-n", "openebs", "--all").Run()
	Expect(err).To(BeNil(), "failed to delete the deployment components")
	if err != nil {
		fmt.Println(err)
	}

	//Removing OpenEBS NDM DaemonSet
	By("Deleting OpenEBS NDM DaemonSet")
	err = exec.Command("kubectl", "delete", "ds", "-n", "openebs", "--all").Run()
	Expect(err).To(BeNil(), "failed to delete the daemonset")
	if err != nil {
		fmt.Println(err)
	}

	//Removing openebs crds
	By("Deleting openebs crds")
	err = exec.Command("kubectl", "delete", "crds", "-n", "openebs", "--all").Run()
	Expect(err).To(BeNil(), "failed to delete all crds")
	if err != nil {
		fmt.Println(err)
	}
	time.Sleep(5 * time.Second)

	//Removing litmus and openebs namespaces
	By("Deleting litmus and openebs namespaces")
	err = exec.Command("kubectl", "delete", "ns", "litmus", "openebs").Run()
	Expect(err).To(BeNil(), "failed to delete all crds")
	if err != nil {
		fmt.Println(err)
	}
	time.Sleep(2 * time.Second)

})
