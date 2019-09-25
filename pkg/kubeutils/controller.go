package kubeutils

import (
	"github.com/golang/glog"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"
	"os"
	"zdyfop/pkg/apis/zdyf/v1alpha1"
	zdyfapiv1alpha1 "zdyfop/pkg/client/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type kubeClient struct {
	zdyfapiClient *Clientset
}

type Clientset struct {
	*discovery.DiscoveryClient
	zdyfapiV1alpha1 *zdyfapiv1alpha1.ZdyfapiV1alpha1Client
}

func (c *Clientset) ZdyfapiV1alpha1() zdyfapiv1alpha1.ZdyfapiV1alpha1Interface {
	return c.zdyfapiV1alpha1
}


type KubeClient interface {
	Get(ns, name string) (*v1alpha1.Zdyfapi, error)
	Update(ns, name string, zdyfapi *v1alpha1.Zdyfapi) (*v1alpha1.Zdyfapi, error)
}

func NewKubeClient(masterUrl, path string) KubeClient {
	var cfg *rest.Config
	var err error
	if path != "" {
		cfg, err = clientcmd.BuildConfigFromFlags(masterUrl, path)
		if err != nil {
			glog.Errorf("Failed to get cluster config with error: %v\n", err)
			os.Exit(1)
		}
	} else {
		cfg, err = rest.InClusterConfig()
		if err != nil {
			glog.Errorf("Failed to get cluster config with error: %v\n", err)
			os.Exit(1)
		}
	}
	client, err := NewForConfig(cfg)
	if err != nil {
		glog.Errorf("Failed to create client with error: %v\n", err)
		os.Exit(1)
	}
	return &kubeClient{client}
}

func (c *kubeClient) Get(ns, name string) (*v1alpha1.Zdyfapi, error) {
	zdyfapi, err := c.zdyfapiClient.ZdyfapiV1alpha1().ZdyfApis(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		glog.Errorf("get hdfs-operator error,err=%+v", err)
		return nil, err
	}
	return zdyfapi, nil
}
func (c *kubeClient) Update(ns, name string, zdyf *v1alpha1.Zdyfapi) (*v1alpha1.Zdyfapi, error) {
	zdyfapi, err := c.zdyfapiClient.ZdyfapiV1alpha1().ZdyfApis(ns).Update(zdyf)
	return zdyfapi, err

}

// NewForConfig creates a new Clientset for the given config.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.zdyfapiV1alpha1, err = zdyfapiv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}
