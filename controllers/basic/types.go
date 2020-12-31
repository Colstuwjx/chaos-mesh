package basic

import (
	"context"
	"fmt"

	"github.com/chaos-mesh/chaos-mesh/api/v1alpha1"
	pkgCfg "github.com/chaos-mesh/chaos-mesh/pkg/config"
	ctx "github.com/chaos-mesh/chaos-mesh/pkg/router/context"
	"github.com/chaos-mesh/chaos-mesh/pkg/router/endpoint"
	"github.com/chaos-mesh/chaos-mesh/pkg/selector"
	v1 "k8s.io/api/core/v1"

	ctrl "sigs.k8s.io/controller-runtime"
)

// Reconciler is basic chaos reconciler implementation
type Reconciler struct {
	endpoint.Endpoint
	ctx.Context
	Config *pkgCfg.ChaosControllerConfig
	Lister selector.ChaosTargetLister
}

// NewReconciler would create Reconciler for common chaos
func NewReconciler(req ctrl.Request, e endpoint.Endpoint, ctx ctx.Context, cfg *pkgCfg.ChaosControllerConfig) *Reconciler {
	ctx.Log = ctx.Log.WithName(req.NamespacedName.String())
	lister := selector.NewChaosTargetLister(
		ctx.Client,
		ctx.Reader,
		cfg.ClusterScoped,
		cfg.TargetNamespace,
		cfg.AllowedNamespaces,
		cfg.IgnoredNamespaces,
	)
	return &Reconciler{
		Endpoint: e,
		Context:  ctx,
		Config:   cfg,
		Lister:   lister,
	}
}

// SelectChaosTargets help reconciler resolve chaos targets
func (r *Reconciler) SelectChaosTargets(ctx context.Context, chaos v1alpha1.InnerObject) (chaosTargets []*v1alpha1.InnerChaosTarget, err error) {
	chaosTargetSelectors := r.Selectors(chaos)
	for _, s := range chaosTargetSelectors {
		var (
			pods []v1.Pod
		)
		pods, err = r.Lister.SelectAndFilterPods(ctx, s)
		if err != nil {
			err = fmt.Errorf("failed to select and generate pods, %s", err.Error())
			return
		}
		var (
			dnsSvc *v1.Service
		)
		dnsSvc, err = r.Lister.GetService(ctx, "", r.Config.Namespace, r.Config.DNSServiceName)
		if err != nil {
			err = fmt.Errorf("failed to get service, %s", err.Error())
			return
		}
		chaosTargets = append(chaosTargets, &v1alpha1.InnerChaosTarget{
			Pods:       pods,
			DNSService: dnsSvc,
		})
	}
	return
}
