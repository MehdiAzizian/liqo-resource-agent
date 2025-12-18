package publisher

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	rearv1alpha1 "github.com/mehdiazizian/liqo-resource-agent/api/v1alpha1"
)

func buildPublisherScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("failed adding client scheme: %v", err)
	}
	if err := rearv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed adding api scheme: %v", err)
	}
	return scheme
}

func newReservationObject(name, requester, target string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "broker.fluidos.eu",
		Version: "v1alpha1",
		Kind:    "Reservation",
	})
	obj.SetName(name)
	obj.Object["spec"] = map[string]interface{}{
		"requesterID":     requester,
		"targetClusterID": target,
		"requestedResources": map[string]interface{}{
			"cpu":    "4",
			"memory": "8Gi",
		},
	}
	obj.Object["status"] = map[string]interface{}{
		"phase":     "Reserved",
		"expiresAt": time.Now().UTC().Format(time.RFC3339),
	}
	return obj
}

func TestUpsertRequesterInstructionCreatesCR(t *testing.T) {
	scheme := buildPublisherScheme(t)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	watcher := &ReservationWatcher{
		LocalClient:          fakeClient,
		InstructionNamespace: "default",
	}

	reservation := newReservationObject("res-1", "rome", "paris")
	status := reservation.Object["status"].(map[string]interface{})
	expiresAt := status["expiresAt"].(string)

	if err := watcher.upsertRequesterInstruction(context.Background(), reservation, "paris", "4", "8Gi", expiresAt); err != nil {
		t.Fatalf("upsert requester instruction failed: %v", err)
	}

	instruction := &rearv1alpha1.ReservationInstruction{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "res-1", Namespace: "default"}, instruction); err != nil {
		t.Fatalf("instruction not created: %v", err)
	}

	if instruction.Spec.TargetClusterID != "paris" {
		t.Fatalf("expected target paris, got %s", instruction.Spec.TargetClusterID)
	}
}

func TestUpsertProviderInstructionCreatesCR(t *testing.T) {
	scheme := buildPublisherScheme(t)
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	watcher := &ReservationWatcher{
		LocalClient:          fakeClient,
		InstructionNamespace: "default",
	}

	reservation := newReservationObject("res-2", "rome", "paris")
	status := reservation.Object["status"].(map[string]interface{})
	expiresAt := status["expiresAt"].(string)

	if err := watcher.upsertProviderInstruction(context.Background(), reservation, "rome", "4", "8Gi", expiresAt); err != nil {
		t.Fatalf("upsert provider instruction failed: %v", err)
	}

	providerInstruction := &rearv1alpha1.ProviderInstruction{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: "res-2-provider", Namespace: "default"}, providerInstruction); err != nil {
		t.Fatalf("provider instruction not created: %v", err)
	}

	if providerInstruction.Spec.RequesterClusterID != "rome" {
		t.Fatalf("expected requester rome, got %s", providerInstruction.Spec.RequesterClusterID)
	}

	if providerInstruction.Status.Enforced {
		t.Fatalf("instruction should start unenforced")
	}
}
