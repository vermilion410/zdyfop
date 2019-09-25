package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ZdyfapiSpec defines the desired state of Zdyfapi
// +k8s:openapi-gen=true

type ZdyfapiSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	//Size int32 `json:"size"`
	NameReplicas   *int32 `json:"namereplicas"`
	NameImage      string `json:"nameimage"`
	NameDepName    string `json:"namedepname"`
	NameDeppodName string `json:"namedeppodname"`
	//LabelSelector map[string]string `json:"labelselector"`

	NameEnvsName   string `json:"nameenvsname"`
	NameEnvsValue  string `json:"nameenvsvalue"`
	NameEnvsName2  string `json:"nameenvsname2"`
	NameEnvsValue2 string `json:"nameenvsvalue2"`
	NameEnvsName3  string `json:"nameenvsname3"`
	NameEnvsName4  string `json:"nameenvsname4"`
	NameEnvsValue3 string `json:"nameenvsvalue3"`
	NameEnvsValue4 string `json:"nameenvsvalue4"`

	NameVolumeMountsName string `json:"namevolumemountsname"`
	NameVolumeMountsPath string `json:"namevolumemountspath"`

	NameContainerPortName1 string `json:"namecontainerportname1"`
	NameContainerPort1     int32  `json:"namecontainerport1"`
	NameContainerPortName2 string `json:"namecontainerportname2"`
	NameContainerPort2     int32  `json:"namecontainerport2"`

	NameVolumesName string `json:"namevolumesname"`
	NamePVCName     string `json:"namepvcname"`
	NamePVName      string `json:"namepvname"`
	NamePvStorage   string `json:"namepvstorage"`

	NameSports     []corev1.ServicePort `json:"namesports,omitempty"`
	NamePVCStorage string               `json:"namepvcstorage"`

	NameSCName string `json:"namescname"`

	DataReplicas *int32 `json:"datareplicas"`
	DataImage    string `json:"dataimage"`
	ServiceName  string `json:"servicename"`

	DataEnvsName  string `json:"dataenvsname"`
	DataEnvsValue string `json:"dataenvsvalue"`

	DataVolumeMountsPath string `json:"datavolumemountspath"`
	DataSport            int32  `json:"datasport"`
	DataPVCName1         string `json:"datapvcname1"`
	DataPVCName2         string `json:"datapvcname2"`
	DataPVCStorage1      string `json:"datapvcstorage1"`
	DataPVCStorage2      string `json:"datapvcstorage2"`
	DataSC               string `json:"datasc"`
	DataPVStorage1       string `json:"datapvstorage1"`
	DataPVStorage2       string `json:"datapvstorage2"`
	DataPVName1          string `json:"datapvname1"`
	DataPVName2          string `json:"datapvname2"`
	DataPVName3          string `json:"datapvname3"`
	DataPVName4          string `json:"datapvname4"`
	DataPVName5          string `json:"datapvname5"`
}

// ZdyfapiStatus defines the observed state of Zdyfapi
// +k8s:openapi-gen=true
type ZdyfapiStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	//appsv1.DeploymentStatus `json:",inline"`
	//Nodes                   []string `json:"nodes"`
	appsv1.DeploymentStatus `json:",inline"`
	// pods      []PodStatus      `json:"pods,omitempty"`
	// Services  []ServiceStatus  `json:"services,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// Zdyfapi is the Schema for the zdyfapis API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Zdyfapi struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZdyfapiSpec   `json:"spec,omitempty"`
	Status ZdyfapiStatus `json:"status,omitempty"`
}

type ZdyfapiService struct {
	// Type is the type of the service
	Type corev1.ServiceType `json:"type"`
	// LoadBalancerIP is an optional load balancer IP for the service
	LoadBalancerIP string `json:"loadBalancerIP"`
	// Labels are extra labels for the service
	Labels map[string]string `json:"labels"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ZdyfapiList contains a list of Zdyfapi
type ZdyfapiList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Zdyfapi `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Zdyfapi{}, &ZdyfapiList{})
}

