package status

import (
	"context"
	hazelcastv1alpha1 "github.com/hazelcast/hazelcast-enterprise-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type OptionsBuilder struct {
	phase      hazelcastv1alpha1.Phase
	retryAfter int
	err        error
}

func FailedPhase(err error) OptionsBuilder {
	return OptionsBuilder{
		phase: hazelcastv1alpha1.Failed,
		err:   err,
	}
}

func PendingPhase(retryAfter int) OptionsBuilder {
	return OptionsBuilder{
		phase:      hazelcastv1alpha1.Pending,
		retryAfter: retryAfter,
	}
}

func RunningPhase() OptionsBuilder {
	return OptionsBuilder{
		phase: hazelcastv1alpha1.Running,
	}
}

// Update takes the options provided by the given OptionsBuilder, applies them all and then updates the Hazelcast resource
func Update(statusWriter client.StatusWriter, h *hazelcastv1alpha1.Hazelcast, options OptionsBuilder) (ctrl.Result, error) {
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
