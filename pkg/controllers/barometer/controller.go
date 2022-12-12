package barometer

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/ellistarn/barometer/pkg/apis/v1alpha1"
	"github.com/samber/lo"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/logging"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func NewController(kubeClient client.Client) *Controller {
	nodeName, ok := os.LookupEnv("NODE_NAME")
	if !ok {
		panic("NODE_NAME not found, handle this better")
	}
	return &Controller{
		kubeClient:   kubeClient,
		nodeName:     nodeName,
		pollInterval: 10 * time.Second,
	}
}

type Controller struct {
	kubeClient   client.Client
	nodeName     string
	pollInterval time.Duration
}

func (c *Controller) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	barometer := &v1alpha1.Barometer{}
	if err := c.kubeClient.Get(ctx, req.NamespacedName, barometer); err != nil {
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}
	podList := v1.PodList{}
	if err := c.kubeClient.List(ctx, &podList,
		client.InNamespace(barometer.Namespace),
		client.MatchingLabels(barometer.Spec.Selector),
		client.MatchingFields{"spec.nodeName": c.nodeName},
	); err != nil {
		return reconcile.Result{}, err
	}

	pressure := map[string]*v1alpha1.PSI{}

	// Remove not pods
	for key := range barometer.Status.Pressure {
		pod := &v1.Pod{}
		err := c.kubeClient.Get(ctx, types.NamespacedName{Namespace: barometer.Namespace, Name: strings.Split(key, "/")[0]}, pod)
		if client.IgnoreNotFound(err) != nil {
			return reconcile.Result{}, err
		}
		if errors.IsNotFound(err) {
			pressure[key] = nil // Zero out pods that aren't found
		}
	}
	// Add pressure for known pods
	for _, pod := range podList.Items {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if !pod.DeletionTimestamp.IsZero() || containerStatus.ContainerID == "" {
				continue
			}
			psi, err := GetPSI(barometer.Spec.Threshold, &pod, &containerStatus)
			if err != nil {
				return reconcile.Result{}, fmt.Errorf("getting pressure, for %s/%s %w", pod.Name, containerStatus.ContainerID, err)
			}
			pressure[fmt.Sprintf("%s/%s", pod.Name, containerStatus.Name)] = psi
		}
	}
	// Patdate status
	patched := barometer.DeepCopy()
	patched.Status.Pressure = pressure
	if !reflect.DeepEqual(barometer.Status, patched.Status) {
		logging.FromContext(ctx).Info("detected changed")
		if err := c.kubeClient.Status().Patch(ctx, patched, client.MergeFrom(barometer)); err != nil {
			return reconcile.Result{}, client.IgnoreNotFound(err)
		}
	}
	return reconcile.Result{RequeueAfter: c.pollInterval}, nil
}

func Register(ctx context.Context, m manager.Manager) error {
	lo.Must0(m.GetFieldIndexer().IndexField(ctx, &v1.Pod{}, "spec.nodeName", func(o client.Object) []string {
		return []string{o.(*v1.Pod).Spec.NodeName}
	}), "failed to setup pod indexer")

	return controllerruntime.NewControllerManagedBy(m).
		For(&v1alpha1.Barometer{}).
		Complete(NewController(m.GetClient()))
}
