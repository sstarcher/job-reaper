package kube

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"k8s.io/client-go/1.4/kubernetes"
	"k8s.io/client-go/1.4/pkg/api"
	"time"

	"k8s.io/client-go/1.4/pkg/api/v1"
	"k8s.io/client-go/1.4/pkg/labels"

	"k8s.io/client-go/1.4/pkg/selection"
	"k8s.io/client-go/1.4/pkg/util/sets"

	"k8s.io/client-go/1.4/pkg/fields"
	"k8s.io/client-go/1.4/tools/clientcmd"

	"github.com/sstarcher/job-reaper/alert"
	batch "k8s.io/client-go/1.4/pkg/apis/batch/v1"
)

// Client Interface for reaping
type Client interface {
	Reap()
}

type kubeClient struct {
	clientset *kubernetes.Clientset
	failures  int
	alerters  *[]alert.Alert
}

// NewKubeClient for interfacing with kubernetes
func NewKubeClient(masterURL string, failures int, alerters *[]alert.Alert) Client {
	config, err := clientcmd.BuildConfigFromFlags(masterURL, "")
	if err != nil {
		log.Panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panic(err.Error())
	}

	return &kubeClient{
		clientset: clientset,
		failures:  failures,
		alerters:  alerters,
	}
}

func (kube *kubeClient) reap(job batch.Job) {
	data := alert.Data{
		Name:      job.GetName(),
		Namespace: job.GetNamespace(),
		Status:    "Unknown",
		Message:   "",
		Config:    job.GetAnnotations(),
	}

	pod := kube.oldestPod(job)
	if scheduledJobName, ok := pod.GetLabels()["run"]; ok {
		data.Name = scheduledJobName
	}

	if pod.Status.Phase != "" {
		data.Status = string(pod.Status.Phase)
	}

	if len(pod.Status.ContainerStatuses) > 0 { //Container has exited
		terminated := pod.Status.ContainerStatuses[0].State.Terminated
		if terminated != nil {
			data.Message = terminated.Reason // ERRRRR
			data.ExitCode = int(terminated.ExitCode)
			data.StartTime = terminated.StartedAt.Time
			data.EndTime = terminated.FinishedAt.Time
		} else {
			log.Error("Unexpected null for container state")
			log.Error(pod.Status.ContainerStatuses[0])
			log.Error(terminated)
			log.Error(job)
			log.Error(job.Status.Conditions)
			log.Error(pod)
			log.Error(kube.podEvents(pod))
			return
		}
	} else if len(job.Status.Conditions) > 0 { //TODO naive when more than one condition
		condition := job.Status.Conditions[0]
		data.Message = fmt.Sprintf("Pod Missing: %s - %s", condition.Reason, condition.Message)
		if condition.Type == batch.JobComplete {
			data.ExitCode = 0
			data.Status = "Succeeded"
		} else {
			data.ExitCode = 998
		}
		data.StartTime = job.Status.StartTime.Time
		data.EndTime = condition.LastTransitionTime.Time
	} else { //Unfinished Containers or missing
		data.ExitCode = 999
		data.EndTime = time.Now()
	}

	for _, alert := range *kube.alerters {
		err := alert.Send(data)
		if err != nil {
			log.Error(err.Error())
		}
	}

	err := kube.clientset.Batch().Jobs(data.Namespace).Delete(job.GetName(), nil)
	if err != nil {
		log.Error(err.Error())
	}

	pods := kube.jobPods(job)
	for _, pod := range pods.Items {
		err := kube.clientset.Core().Pods(data.Namespace).Delete(pod.GetName(), nil)
		if err != nil {
			log.Error(err.Error())
		}
	}
}

func (kube *kubeClient) jobPods(job batch.Job) *v1.PodList {
	controllerUID := job.Spec.Selector.MatchLabels["controller-uid"]
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement("controller-uid", selection.Equals, sets.NewString(controllerUID))
	if err != nil {
		log.Panic(err.Error())
	}
	selector = selector.Add(*requirement)
	pods, err := kube.clientset.Core().Pods(job.ObjectMeta.Namespace).List(api.ListOptions{LabelSelector: selector})
	if err != nil {
		log.Panic(err.Error())
	}
	return pods
}

func (kube *kubeClient) podEvents(pod v1.Pod) *v1.EventList {
	sel, err := fields.ParseSelector("involvedObject.name=" + pod.ObjectMeta.Name)
	if err != nil {
		log.Panic(err.Error())
	}
	events, err := kube.clientset.Core().Events(pod.ObjectMeta.Namespace).List(api.ListOptions{
		FieldSelector: sel,
	})
	return events
}

func (kube *kubeClient) oldestPod(job batch.Job) v1.Pod {
	pods := kube.jobPods(job)
	time := time.Now()
	var tempPod v1.Pod
	for _, pod := range pods.Items {
		if time.After(pod.ObjectMeta.CreationTimestamp.Time) {
			time = pod.ObjectMeta.CreationTimestamp.Time
			tempPod = pod
		}
	}
	return tempPod
}

func (kube *kubeClient) Reap() {
	namespaces, err := kube.clientset.Core().Namespaces().List(api.ListOptions{})
	if err != nil {
		log.Panic(err.Error())
	}
	for _, namespace := range namespaces.Items {
		log.Debugf("Processing namespace: %s", namespace.ObjectMeta.Name)
		kube.reapNamespace(namespace.ObjectMeta.Name)
	}
}

func (kube *kubeClient) reapNamespace(namespace string) {
	jobs, err := kube.clientset.Batch().Jobs(namespace).List(api.ListOptions{})
	if err != nil {
		log.Panic(err.Error())
	}

	for _, job := range jobs.Items {
		var completions = 1
		if job.Spec.Completions == nil {
			completions = int(*job.Spec.Completions)
		}

		if int(job.Status.Succeeded) >= completions {
			kube.reap(job)
			continue
		}

		if int(job.Status.Failed) > kube.failures && kube.failures > -1 {
			kube.reap(job)
			continue
		}
	}
}
