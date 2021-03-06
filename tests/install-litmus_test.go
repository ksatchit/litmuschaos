package tests

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/mayadata-io/chaos-ci-lib/pkg"
	chaosTypes "github.com/mayadata-io/chaos-ci-lib/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	scheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	chaosClient "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/typed/litmuschaos/v1alpha1"
	restclient "k8s.io/client-go/rest"
)

var (
	kubeconfig string
	config     *restclient.Config
	client     *kubernetes.Clientset
	clientSet  *chaosClient.LitmuschaosV1alpha1Client
	err        error
	out        bytes.Buffer
	stderr     bytes.Buffer
)

func testChaos(t *testing.T) {

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

})

//BDD Tests to Install Litmus
var _ = Describe("BDD of Litmus installation", func() {

	// BDD TEST CASE 1
	Context("Check for the Litmus components", func() {

		It("Should check for creation of Litmus", func() {

			//Installing Litmus
			By("Installing Litmus")
			klog.Info("Installing Litmus")
			err = pkg.DownloadFile("install-litmus.yaml", chaosTypes.InstallLitmus)
			Expect(err).To(BeNil(), "fail to fetch operator yaml file to install litmus")
			klog.Info("Updating Operator Image")
			err = pkg.EditFile("install-litmus.yaml", "image: litmuschaos/chaos-operator:latest", "image: "+pkg.GetEnv("OPERATOR_IMAGE", "litmuschaos/chaos-operator:latest"))
			Expect(err).To(BeNil(), "Failed to update the operator image")
			klog.Info("Updating Runner Image")
			err = pkg.EditKeyValue("install-litmus.yaml", "CHAOS_RUNNER_IMAGE", "value: \"litmuschaos/chaos-runner:latest\"", "value: '"+pkg.GetEnv("RUNNER_IMAGE", "litmuschaos/chaos-runner:latest")+"'")
			Expect(err).To(BeNil(), "Failed to update chaos interval")
			cmd := exec.Command("kubectl", "apply", "-f", "install-litmus.yaml")
			cmd.Stdout = &out
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
				fmt.Println(err)
				Fail("Fail to install litmus")
			}
			fmt.Println("Result: " + out.String())

			//Checking the status of operator
			operator, _ := client.AppsV1().Deployments(pkg.GetEnv("APP_NS", "default")).Get("chaos-operator-ce", metav1.GetOptions{})
			count := 0
			for operator.Status.UnavailableReplicas != 0 {
				if count < 50 {
					fmt.Printf("Unavaliable Count: %v \n", operator.Status.UnavailableReplicas)
					operator, _ = client.AppsV1().Deployments(pkg.GetEnv("APP_NS", "default")).Get("chaos-operator-ce", metav1.GetOptions{})
					time.Sleep(5 * time.Second)
					count++
				} else {
					Fail("Operator is not in Ready state Time Out")
				}
			}
			klog.Info("Chaos Operator created successfully")
			klog.Info("Litmus installed successfully")
		})
	})

})
