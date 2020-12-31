// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package httpchaos

import (
	"context"
	"errors"

	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	"github.com/chaos-mesh/chaos-mesh/pkg/events"
	"github.com/chaos-mesh/chaos-mesh/pkg/finalizer"
	"github.com/chaos-mesh/chaos-mesh/pkg/router"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	end "github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
)

type endpoint struct {
	ctx.Context
}

func (r *endpoint) Apply(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject, chaosTargets []*v1alpha1.InnerChaosTarget) error {
	httpChaos, ok := chaos.(*v1alpha1.HTTPChaos)
	if !ok {
		err := errors.New("chaos is not HttpChaos")
		r.Log.Error(err, "chaos is not HttpChaos", "chaos", chaos)
		return err
	}
	if len(chaosTargets) != 1 {
		err := errors.New("unexpected chaos target for HttpChaos")
		r.Log.Error(err, "invalid chaos target", "chaos", chaos)
		return err
	}
	if err := r.applyAllPods(ctx, chaosTargets[0].Pods, httpChaos); err != nil {
		r.Log.Error(err, "failed to apply chaos on all pods")
		return err
	}
	return nil
}

func (r *endpoint) Recover(ctx context.Context, req ctrl.Request, chaos v1alpha1.InnerObject) error {
	httpChaos, ok := chaos.(*v1alpha1.HTTPChaos)
	if !ok {
		err := errors.New("chaos is not HttpChaos")
		r.Log.Error(err, "chaos is not HttpChaos", "chaos", chaos)
		return err
	}
	r.Event(httpChaos, v1.EventTypeNormal, events.ChaosRecovered, "")
	return nil
}

func (r *endpoint) Object() v1alpha1.InnerObject {
	return &v1alpha1.HTTPChaos{}
}

// Selectors would return the chaos target selectors
func (r *endpoint) Selectors(chaos v1alpha1.InnerObject) (selectors []v1alpha1.InnerSelector) {
	httpchaos, ok := chaos.(*v1alpha1.HTTPChaos)
	if !ok {
		r.Log.Error("chaos is not HTTPChaos", "chaos", chaos)
		return
	}
	selectors = append(selectors, &httpchaos.Spec)
	return selectors
}

func (r *endpoint) applyAllPods(ctx context.Context, pods []v1.Pod, chaos *v1alpha1.HTTPChaos) error {
	g := errgroup.Group{}
	for index := range pods {
		pod := &pods[index]

		key, err := cache.MetaNamespaceKeyFunc(pod)
		if err != nil {
			return err
		}
		chaos.Finalizers = finalizer.InsertFinalizer(chaos.Finalizers, key)

		g.Go(func() error {
			return r.applyPod(ctx, pod, chaos)
		})
	}

	return g.Wait()
}

func (r *endpoint) applyPod(ctx context.Context, pod *v1.Pod, chaos *v1alpha1.HTTPChaos) error {
	//TODO: The way to connect with sidecar need be discussed & It will work after the sidecar add to the repo.
	r.Log.Info("Try to inject Http chaos on pod", "namespace", pod.Namespace, "name", pod.Name)
	return nil
}

func init() {
	router.Register("httpchaos", &v1alpha1.HTTPChaos{}, func(obj runtime.Object) bool {
		return true
	}, func(ctx ctx.Context) end.Endpoint {
		return &endpoint{
			Context: ctx,
		}
	})
}
