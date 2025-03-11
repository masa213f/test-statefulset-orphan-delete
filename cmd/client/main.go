package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"github.com/masa213f/test-statefulset-orphan-delete/pkg/constant"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func create(client v1.StatefulSetInterface, pvcLabelValue string) {
	// fmt.Println("Creating...")
	sts := constant.GetAppsV1StatefulSet()
	if pvcLabelValue != "" {
		if sts.Spec.VolumeClaimTemplates[0].Labels == nil {
			sts.Spec.VolumeClaimTemplates[0].Labels = map[string]string{}
		}
		sts.Spec.VolumeClaimTemplates[0].Labels["key"] = pvcLabelValue
	}
	_, err := client.Create(context.Background(), sts, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	// fmt.Printf("Created %q.\n", result.GetObjectMeta().GetName())
}

func createSSA(client v1.StatefulSetInterface, pvcLabelValue string) {
	// fmt.Println("Creating...")
	sts := constant.GetAppsV1StatefulSetAC()
	if pvcLabelValue != "" {
		if sts.Spec.VolumeClaimTemplates[0].Labels == nil {
			sts.Spec.VolumeClaimTemplates[0].Labels = map[string]string{}
		}
		sts.Spec.VolumeClaimTemplates[0].Labels["key"] = pvcLabelValue
	}
	_, err := client.Apply(context.Background(), sts, metav1.ApplyOptions{
		FieldManager: constant.FieldManagerClient,
		Force:        true,
	})
	if err != nil {
		panic(err)
	}
	// fmt.Printf("Created %q.\n", result.GetObjectMeta().GetName())
}

func delete(client v1.StatefulSetInterface, deletePolicy metav1.DeletionPropagation) {
	// fmt.Printf("Deleting... (%s)\n", deletePolicy)
	if err := client.Delete(context.Background(), constant.StatefulSetName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}
	for {
		_, err := client.Get(context.Background(), constant.StatefulSetName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			break
		} else if err != nil {
			panic(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	// fmt.Printf("Deleted.\n")
}

func wait(client v1.StatefulSetInterface) {
	// fmt.Println("Waiting...")
	for {
		sts, err := client.Get(context.Background(), constant.StatefulSetName, metav1.GetOptions{})
		if err != nil {
			panic(err)
		}
		if sts.Generation == sts.Status.ObservedGeneration &&
			sts.Status.Replicas == sts.Status.ReadyReplicas {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	// fmt.Println("Ready!")
}

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	clientset.RESTClient().Get()
	stsClient := clientset.AppsV1().StatefulSets(constant.Namespace)

	cmd := flag.Args()
	fmt.Println(flag.Args())
	for _, c := range cmd {
		switch c {
		case "create":
			create(stsClient, "")
		case "createWithLabel":
			create(stsClient, "value")
		case "createSSA":
			createSSA(stsClient, "")
		case "createSSAWithLabel":
			createSSA(stsClient, "value")
		case "delete":
			delete(stsClient, metav1.DeletePropagationForeground)
		case "orphanDelete":
			delete(stsClient, metav1.DeletePropagationOrphan)
		case "wait":
			wait(stsClient)
		default:
			fmt.Printf("unknown command: %s\n", c)
		}
	}
}
