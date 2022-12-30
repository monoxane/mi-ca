package controller

import (
	"encoding/json"
	"fmt"
	"mi-ca/internal/env"
	"mi-ca/internal/kubernetes"
	"mi-ca/internal/logging"
	"mi-ca/internal/model"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/openlyinc/pointy"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	done chan interface{}
	log  logging.Logger
	conn *websocket.Conn
)

func Connect() {
	log = logging.Log.With().Str("module", "controller").Logger()
	done = make(chan interface{})

	socketUrl := fmt.Sprintf("%s/v1/ca", env.C2_SERVER)
	var err error
	conn, _, err = websocket.DefaultDialer.Dial(socketUrl, nil)
	if err != nil {
		log.Error().Msgf("Error connecting to Websocket Server:", err)
		os.Exit(1)
	}
	defer conn.Close()
	go receiveHandler(conn)

	initialUpdate := makeUpdateFromKubeState()
	initialUpdate.Cause = "boot"
	conn.WriteJSON(initialUpdate)

	go WatchDeployments()

	for {
		time.Sleep(time.Duration(1) * time.Second * 60)

		err := conn.WriteJSON(makeUpdateFromKubeState())
		if err != nil {
			log.Error().Msgf("Error during writing to websocket:", err)
			os.Exit(1)
		}
	}
}

func receiveHandler(connection *websocket.Conn) {
	defer close(done)
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			log.Error().Msgf("Error in receive:", err)
			os.Exit(1)
			return
		}
		log.Printf("Received: %s", msg)

		var update model.Cluster

		unmarshallErr := json.Unmarshal(msg, &update)
		if unmarshallErr != nil {
			log.Error().Msgf("unable to unmarshal command:", err)
		} else {
			if update.Project == "none" {
				stopWorkers()
			} else {
				setProject(update.Cause, update.Project, update.Workers, update.Concurrency)
			}
		}
	}
}

func makeUpdateFromKubeState() *model.Cluster {
	dep, err := kubernetes.GetDeployment("mi-worker")
	if err != nil {
		log.Warn().Err(err).Msg("kubernetes returned an error getting the deployment")
		return &model.Cluster{
			Name:        env.CLUSTER_NAME,
			Owners:      env.CLUSTER_OWNERS,
			Project:     "none",
			Workers:     0,
			Concurrency: 0,
		}
	}

	concurrency, _ := strconv.Atoi(dep.ObjectMeta.Labels["app.kubernetes.io/concurrency"])

	return &model.Cluster{
		Name:          env.CLUSTER_NAME,
		Owners:        env.CLUSTER_OWNERS,
		Project:       dep.ObjectMeta.Labels["app.kubernetes.io/project"],
		Workers:       int(dep.Status.Replicas),
		ActiveWorkers: int(dep.Status.AvailableReplicas),
		Concurrency:   concurrency,
		Cause:         dep.ObjectMeta.Labels["app.kubernetes.io/cause"],
	}
}

func WatchDeployments() {
	events, _ := kubernetes.WatchDeployments()

	for event := range events {
		deployment, ok := event.Object.(*appsv1.Deployment)
		if !ok {
			continue
		}

		log.Debug().Interface("type", event.Type).Str("deployment", deployment.Name).Msg("recieved update for deployment")
		conn.WriteJSON(makeUpdateFromKubeState())
	}

	go WatchDeployments()
}

func stopWorkers() {
	kubernetes.DeleteDeployment("mi-worker")
}

func setProject(cause, project string, workers, concurrency int) {
	deploymentTemplate := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mi-worker",
			Namespace: env.KUBERNETES_NAMESPACE,
			Labels: map[string]string{
				"app.kubernetes.io/name":        "mi-worker",
				"app.kubernetes.io/concurrency": fmt.Sprintf("%d", concurrency),
				"app.kubernetes.io/project":     project,
				"app.kubernetes.io/cause":       cause,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointy.Int32(int32(workers)),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name": "mi-worker",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name": "mi-worker",
					},
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: pointy.Int64(60),
					Containers: []corev1.Container{
						{
							Name:  "mi-worker",
							Image: fmt.Sprintf("atdr.meo.ws/archiveteam/%s-grab", project),
							Args: []string{
								"--concurrent",
								fmt.Sprintf("%d", concurrency),
								"--disable-web-server",
								"MagnusInstitute",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "SHARED_RSYNC_THREADS",
									Value: "2",
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									MountPath: "/grab/data",
									Name:      "cache-volume",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "cache-volume",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	_, err := kubernetes.UpdateDeployment(&deploymentTemplate)
	if err != nil {
		log.Warn().Err(err).Msg("kubernetes returned an error applying the deployment")
	}
}
