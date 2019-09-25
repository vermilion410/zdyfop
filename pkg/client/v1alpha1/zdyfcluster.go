package v1alpha1

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"zdyfop/pkg/apis/zdyf/v1alpha1"
)



// TidbClustersGetter has a method to return a TidbClusterInterface.
// A group's client should implement this interface.
type ZdyfapisGetter interface {
	ZdyfApis(namespace string) ZdyfapiInterface
}

type ZdyfapiInterface interface {
	Get(*v1alpha1.Zdyfapi) (*v1alpha1.Zdyfapi, error)
	Update(*v1alpha1.Zdyfapi) (*v1alpha1.Zdyfapi, error)
	Create(*v1alpha1.Zdyfapi) (*v1alpha1.Zdyfapi, error)
	List(*v1alpha1.Zdyfapi) (*v1alpha1.Zdyfapi, error)
}

type Zdyfapis struct {
	client rest.Interface
	ns     string
}

func newZdyfapis(c *ZdyfapiV1alpha1Client, namespace string) *Zdyfapis {
	return &Zdyfapis{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the tidbCluster, and returns the corresponding tidbCluster object, and an error if there is any.
func (c *Zdyfapis) Get(name string, options v1.GetOptions) (result *v1alpha1.Zdyfapi, err error) {
	result = &v1alpha1.Zdyfapi{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("zdyfapis").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of TidbClusters that match those selectors.
func (c *Zdyfapis) List(opts v1.ListOptions) (result *v1alpha1.ZdyfapiList, err error) {
	result = &v1alpha1.ZdyfapiList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("zdyfapis").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Create takes the representation of a tidbCluster and creates it.  Returns the server's representation of the tidbCluster, and an error, if there is any.
func (c *Zdyfapis) Create(tidbCluster *v1alpha1.Zdyfapi) (result *v1alpha1.Zdyfapi, err error) {
	result = &v1alpha1.Zdyfapi{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("zdyfapis").
		Body(tidbCluster).
		Do().
		Into(result)
	return
}

// Update takes the representation of a tidbCluster and updates it. Returns the server's representation of the tidbCluster, and an error, if there is any.
func (c *Zdyfapis) Update(tidbCluster *v1alpha1.Zdyfapi) (result *v1alpha1.Zdyfapi, err error) {
	result = &v1alpha1.Zdyfapi{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("zdyfapis").
		Name(tidbCluster.Name).
		Body(tidbCluster).
		Do().
		Into(result)
	return
}