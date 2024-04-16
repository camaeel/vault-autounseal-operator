package kubeclient

import (
	"context"
	"fmt"
	"time"

	"log/slog"

	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

func LeaderElection(ctx context.Context, cfg *config.Config, workload func(ctx2 context.Context), cancel func()) {
	id := uuid.New().String()
	slog.Info(fmt.Sprintf("Starting leader election with id: %s", id))

	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      cfg.LeaseName,
			Namespace: cfg.LeaseNamespace,
		},
		Client: cfg.K8sClient.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: id,
		},
	}
	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock: lock,
		// IMPORTANT: you MUST ensure that any code you have that
		// is protected by the lease must terminate **before**
		// you call cancel. Otherwise, you could have a background
		// loop still running and another process could
		// get elected before your background loop finished, violating
		// the stated goal of the lease.
		ReleaseOnCancel: true,
		LeaseDuration:   60 * time.Second,
		RenewDeadline:   20 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				// we're notified when we start - this is where you would
				// usually put your code
				workload(ctx)
			},
			OnStoppedLeading: func() {
				// we can do cleanup here
				slog.Info(fmt.Sprintf("leader lost: %s", id))
				//os.Exit(0)
				cancel()
			},
			OnNewLeader: func(identity string) {
				// we're notified when new leader elected
				if identity == id {
					// I just got the lock
					return
				}
				slog.Info(fmt.Sprintf("new leader elected: %s", identity))
			},
		},
	})
}
