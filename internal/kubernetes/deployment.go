package kubernetes

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func GetDeployment(name string) (*appsv1.Deployment, error) {
	deployment, err := deploymentsClient.Get(context.Background(), name, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	return deployment, nil
}

func UpdateDeployment(deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	applied, updateErr := deploymentsClient.Update(context.Background(), deployment, metav1.UpdateOptions{})

	if updateErr != nil {
		var createErr error
		applied, createErr = deploymentsClient.Create(context.Background(), deployment, metav1.CreateOptions{})

		if createErr != nil {
			return nil, fmt.Errorf("update err %s, create err %s", updateErr, createErr)
		}
	}

	return applied, nil
}

func DeleteDeployment(name string) error {
	err := deploymentsClient.Delete(context.Background(), name, metav1.DeleteOptions{})

	if err != nil {
		return err
	}

	return nil
}

func WatchDeployments() (<-chan watch.Event, func()) {
	watch, err := deploymentsClient.Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error().Err(err).Msg("failed to watch deployments")
		return nil, nil
	}

	return watch.ResultChan(), watch.Stop
}
