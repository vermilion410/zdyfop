package v1alpha1

import (
	"k8s.io/client-go/rest"
	"zdyfop/pkg/apis/zdyf/v1alpha1"
	serializer "k8s.io/apimachinery/pkg/runtime/serializer"
)



type ZdyfapiV1alpha1Interface interface {
	RESTClient() rest.Interface
	ZdyfapisGetter
}


type ZdyfapiV1alpha1Client struct {
	restClient rest.Interface
}

func (c *ZdyfapiV1alpha1Client) ZdyfApis(namespace string) ZdyfapiInterface {
	panic("implement me")
}

func (c *ZdyfapiV1alpha1Client) Zdyfapis(namespace string) ZdyfapiInterface {
	return newZdyfapis(c, namespace)
}


// NewForConfig creates a new PingcapV1alpha1Client for the given config.
func NewForConfig(c *rest.Config) (*ZdyfapiV1alpha1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return &ZdyfapiV1alpha1Client{client}, nil
}

func NewForConfigOrDie(c *rest.Config) *ZdyfapiV1alpha1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}
// New creates a new PingcapV1alpha1Client for the given RESTClient.
func New(c rest.Interface) *ZdyfapiV1alpha1Client {
	return &ZdyfapiV1alpha1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1alpha1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}


// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *ZdyfapiV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
