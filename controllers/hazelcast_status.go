package controllers

import (
	"context"
	hazelcastv1alpha1 "github.com/hazelcast/hazelcast-enterprise-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type optionsBuilder struct {
	phase      hazelcastv1alpha1.Phase
	retryAfter int
	err        error
}

func failedPhase(err error) optionsBuilder {
	return optionsBuilder{
		phase: hazelcastv1alpha1.Failed,
		err:   err,
	}
}

func pendingPhase(retryAfter int) optionsBuilder {
	return optionsBuilder{
		phase:      hazelcastv1alpha1.Pending,
		retryAfter: retryAfter,
	}
}

func runningPhase() optionsBuilder {
	return optionsBuilder{
		phase: hazelcastv1alpha1.Running,
	}
}

// update takes the options provided by the given optionsBuilder, applies them all and then updates the Hazelcast resource
func update(statusWriter client.StatusWriter, h *hazelcastv1alpha1.Hazelcast, options optionsBuilder) (ctrl.Result, error) {
	h.Status = hazelcastv1alpha1.HazelcastStatus{Phase: options.phase}
	if err := statusWriter.Update(context.TODO(), h); err != nil {
		return ctrl.Result{}, err
	}
	if options.phase == hazelcastv1alpha1.Failed {
		return ctrl.Result{}, options.err
	}
	if options.phase == hazelcastv1alpha1.Pending {
		return ctrl.Result{Requeue: true, RequeueAfter: time.Second * time.Duration(options.retryAfter)}, nil
	}
	return ctrl.Result{}, nil
}
