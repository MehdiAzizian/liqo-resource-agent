#!/bin/bash
set -e

echo "=== Testing Resource Agent Locally ==="
echo ""

echo "Step 1: Creating kind cluster..."
kind create cluster --name liqo-test 2>/dev/null || echo "Cluster already exists"

echo ""
echo "Step 2: Installing CRDs..."
make install

echo ""
echo "Step 3: Applying sample Advertisement..."
kubectl apply -f config/samples/rear_v1alpha1_advertisement.yaml

echo ""
echo "Step 4: Waiting for controller to reconcile..."
sleep 5

echo ""
echo "Step 5: Checking Advertisement status..."
kubectl get advertisements

echo ""
echo "Step 6: Detailed view..."
kubectl describe advertisement cluster-advertisement

echo ""
echo "=== Test Complete ==="
echo "Run 'make run' in another terminal to start the controller"
