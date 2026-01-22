package main

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	boilerrv1alpha1 "github.com/CraightonH/boilerr/api/v1alpha1"
	"github.com/CraightonH/boilerr/internal/resources"
)

func main() {
	// Define GameDefinition (what operator bundles)
	gameDef := &boilerrv1alpha1.GameDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "valheim",
		},
		Spec: boilerrv1alpha1.GameDefinitionSpec{
			AppId:   896660,
			Command: "./valheim_server.x86_64",
			Args: []string{
				"-nographics",
				"-batchmode",
				"-port", "{{.Config.port}}",
				"-name", "{{.Config.serverName}}",
				"-world", "{{.Config.worldName}}",
				"-password", "{{.Config.password}}",
			},
			Ports: []boilerrv1alpha1.ServerPort{
				{Name: "game", ContainerPort: 2456, Protocol: corev1.ProtocolUDP},
				{Name: "query", ContainerPort: 2457, Protocol: corev1.ProtocolUDP},
			},
			DefaultStorage: "20Gi",
		},
	}

	// Define SteamServer (what user creates)
	server := &boilerrv1alpha1.SteamServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-valheim",
			Namespace: "default",
		},
		Spec: boilerrv1alpha1.SteamServerSpec{
			GameDefinition: "valheim",
			Config: map[string]boilerrv1alpha1.ConfigValue{
				"serverName": {Value: "Vikings Valhalla"},
				"worldName":  {Value: "Midgard"},
				"password": {
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{Name: "valheim-secrets"},
						Key:                  "password",
					},
				},
				"port": {Value: "2456"},
			},
			Storage: &boilerrv1alpha1.StorageSpec{
				Size: resource.MustParse("30Gi"),
			},
			ServiceType: corev1.ServiceTypeLoadBalancer,
		},
	}

	fmt.Println("=== INPUT: User's SteamServer CR ===")
	printYAML(server)

	fmt.Println("\n=== GENERATED: StatefulSet ===")
	stsBuilder := resources.NewStatefulSetBuilder(server, gameDef)
	sts := stsBuilder.Build()
	printYAML(sts)

	fmt.Println("\n=== GENERATED: Service ===")
	svcBuilder := resources.NewServiceBuilder(server, gameDef)
	svc := svcBuilder.Build()
	printYAML(svc)

	fmt.Println("\n=== GENERATED: PVC ===")
	pvcBuilder := resources.NewPVCBuilder(server, gameDef)
	pvc := pvcBuilder.Build()
	printYAML(pvc)
}

func printYAML(obj interface{}) {
	data, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	yamlData, err := yaml.JSONToYAML(data)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(yamlData))
}
