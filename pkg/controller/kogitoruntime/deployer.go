// Copyright 2020 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kogitoruntime

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"io/ioutil"
	"regexp"
)

const (
	kogitoHome = "/home/kogito"

	serviceAccountName = "kogito-service-viewer"

	envVarExternalURL = "KOGITO_SERVICE_URL"

	// protobufConfigMapSuffix Suffix that is appended to Protobuf ConfigMap name
	protobufConfigMapSuffix  = "protobuf-files"
	downwardAPIVolumeName    = "podinfo"
	downwardAPIVolumeMount   = kogitoHome + "/" + downwardAPIVolumeName
	downwardAPIProtoBufCMKey = "protobufcm"
	protobufSubdir = "/persistence/protobuf/"
	protobufListFileName = "list.json"

	envVarNamespace = "NAMESPACE"
)

var (
	downwardAPIDefaultMode = int32(420)
)

func onGetComparators(comparator compare.ResourceComparator) {
	comparator.SetComparator(
		framework.NewComparatorBuilder().
			WithType(reflect.TypeOf(corev1.ConfigMap{})).
			WithCustomComparator(protoBufConfigMapComparator).
			Build())

	comparator.SetComparator(
		framework.NewComparatorBuilder().
			WithType(reflect.TypeOf(monv1.ServiceMonitor{})).
			WithCustomComparator(framework.CreateServiceMonitorComparator()).
			Build())
}

func onObjectsCreate(cli *client.Client, kogitoService v1alpha1.KogitoService) (resources map[reflect.Type][]resource.KubernetesResource, lists []runtime.Object, err error) {
	resources = make(map[reflect.Type][]resource.KubernetesResource)

	resObjectList, resType, res := createProtoBufConfigMap(cli, kogitoService)
	lists = append(lists, resObjectList)
	resources[resType] = []resource.KubernetesResource{res}
	return
}

func getProtobufData(cli *client.Client, kogitoService v1alpha1.KogitoService) (map[string]string) {
	available, err := services.IsDeploymentAvailable(cli, kogitoService)
	if err != nil {
		log.Errorf("failed to check status of %s, error message: %s", kogitoService.GetName(), err.Error())
		return nil
	}
	if !available {
		log.Debugf("deployment not available yet for %s ", kogitoService.GetName())
		return nil
	}

	// print endpoint
	protobufEndpoint := infrastructure.GetKogitoServiceEndpoint(kogitoService) + protobufSubdir
	log.Debugf("%s Protobuf Endpoint: %s", kogitoService.GetName(), protobufEndpoint)

	// print protobuf list
	protobufListBytes, err := getHTTPFileBytes(protobufEndpoint + protobufListFileName)
	if err != nil {
		log.Errorf("failed to get %s protobuf file list, error message: %s", kogitoService.GetName(), err.Error())
		return nil
	}
	log.Debugf("%s Protobuf List: %s", kogitoService.GetName(), protobufListBytes)

	// create protobuf ConfigMap object
	protobufList := strings.Split(string(protobufListBytes), ",")
	// remove square brackets, commas and quotes from split file names
	r, _ := regexp.Compile("[\",\\[\\]]{1}")
	var protobufFileBytes []byte 
	data := map[string]string{}
	for _, s := range protobufList {
		fileName := r.ReplaceAllString(s, "")
		protobufFileBytes, err = getHTTPFileBytes(protobufEndpoint + fileName)
		if err != nil {
			log.Errorf("failed to get %s, error message: %s", fileName, err.Error())
			return data
		}
		data[fileName] = string(protobufFileBytes)
	}
	return data
}

func createProtoBufConfigMap(cli *client.Client, kogitoService v1alpha1.KogitoService) (runtime.Object, reflect.Type, resource.KubernetesResource) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: kogitoService.GetNamespace(),
			Name:      getProtoBufConfigMapName(kogitoService.GetName()),
			Labels: map[string]string{
				infrastructure.ConfigMapProtoBufEnabledLabelKey: "true",
				framework.LabelAppKey:                           kogitoService.GetName(),
			},
		},
		Data: getProtobufData(cli, kogitoService),
	}
	return &corev1.ConfigMapList{}, reflect.TypeOf(corev1.ConfigMap{}), configMap
}

func protoBufConfigMapComparator(deployed resource.KubernetesResource, requested resource.KubernetesResource) (equal bool) {
	return framework.CreateConfigMapComparator()(deployed, requested)
}

// onDeploymentCreate hooks into the infrastructure package to add additional capabilities/properties to the deployment creation
func onDeploymentCreate(cli *client.Client, deployment *v1.Deployment, kogitoService v1alpha1.KogitoService) error {
	kogitoRuntime := kogitoService.(*v1alpha1.KogitoRuntime)
	// NAMESPACE service discovery
	framework.SetEnvVar(envVarNamespace, kogitoService.GetNamespace(), &deployment.Spec.Template.Spec.Containers[0])
	// external URL
	if kogitoService.GetStatus().GetExternalURI() != "" {
		framework.SetEnvVar(envVarExternalURL, kogitoService.GetStatus().GetExternalURI(), &deployment.Spec.Template.Spec.Containers[0])
	}
	// sa
	deployment.Spec.Template.Spec.ServiceAccountName = serviceAccountName
	// istio
	if kogitoRuntime.Spec.EnableIstio {
		framework.AddIstioInjectSidecarAnnotation(&deployment.Spec.Template.ObjectMeta)
	}
	// protobuf
	applyProtoBufConfigurations(deployment, kogitoService)

	if err := infrastructure.InjectDataIndexURLIntoKogitoRuntimeDeployment(cli, kogitoService.GetNamespace(), deployment); err != nil {
		return err
	}

	if err := infrastructure.InjectJobsServiceURLIntoKogitoRuntimeDeployment(cli, kogitoService.GetNamespace(), deployment); err != nil {
		return err
	}

	return nil
}

// getProtoBufConfigMapName gets the name of the protobuf configMap based the given KogitoRuntime instance
func getProtoBufConfigMapName(serviceName string) string {
	return fmt.Sprintf("%s-%s", serviceName, protobufConfigMapSuffix)
}

// applyProtoBufConfigurations configures the deployment to handle protobuf
func applyProtoBufConfigurations(deployment *v1.Deployment, kogitoService v1alpha1.KogitoService) {
	deployment.Spec.Template.Labels[downwardAPIProtoBufCMKey] = getProtoBufConfigMapName(kogitoService.GetName())
	deployment.Spec.Template.Spec.Volumes = append(
		deployment.Spec.Template.Spec.Volumes,
		corev1.Volume{
			Name: downwardAPIVolumeName,
			VolumeSource: corev1.VolumeSource{
				DownwardAPI: &corev1.DownwardAPIVolumeSource{
					Items: []corev1.DownwardAPIVolumeFile{
						{Path: "name", FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name", APIVersion: "v1"}},
						{Path: downwardAPIProtoBufCMKey, FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['" + downwardAPIProtoBufCMKey + "']", APIVersion: "v1"}},
					},
					DefaultMode: &downwardAPIDefaultMode,
				},
			},
		},
	)
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts =
			append(
				deployment.Spec.Template.Spec.Containers[0].VolumeMounts,
				corev1.VolumeMount{
					Name:      downwardAPIVolumeName,
					MountPath: downwardAPIVolumeMount,
				})
	}
}

func getHTTPFileBytes(fileURL string) ([]byte, error) {
	res, err := http.Get(fileURL)
	if err != nil {
		log.Errorf("failed to download file at %s, error message: %s", fileURL, err.Error())
		return nil, err
	}
	fileBytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	return fileBytes, nil
}
