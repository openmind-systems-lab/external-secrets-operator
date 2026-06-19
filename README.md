# 🔐 External Secrets Operator (ESO) on Docker Desktop Kubernetes

## 📖 Overview

This Proof of Concept demonstrates how **External Secrets Operator (ESO)** synchronizes a secret from an external provider into a Kubernetes Secret and how an application can consume secret updates **without requiring a pod restart**.

---

## ✅ Prerequisites

- 🐳 Docker Desktop (Kubernetes enabled)
- ☸️ kubectl
- ⛵ Helm
- 🐋 Docker

---

## 🏗️ Architecture

![Architecture](media/schema.png)

---

## 🚀 Install External Secrets Operator

Add the Helm repository:

```bash
helm repo add external-secrets https://charts.external-secrets.io
helm repo update
```

Install ESO:

```bash
helm install external-secrets external-secrets/external-secrets \
  -n external-secrets \
  --create-namespace \
  --set installCRDs=true
```

Verify the installation:

```bash
kubectl get pods -n external-secrets
kubectl get crd | grep external-secrets
```

---

## 🔑 Deploy the Fake Provider

Deploy the resources:

```bash
kubectl apply -f eso-test.yaml
```

Verify that the External Secret is synchronized:

```bash
kubectl get externalsecret -n eso-test
```

Verify the generated Kubernetes Secret:

```bash
kubectl get secret demo-k8s-secret \
  -n eso-test \
  -o jsonpath='{.data.DB_PASSWORD}' | base64 -d
```

Expected output:

```text
super-local-password
```

---

## 🛠️ Build the Go Application

Build the Docker image:

```bash
docker build -t eso-logger:local .
```

---

## 🚢 Deploy the Application

Deploy the logger:

```bash
kubectl apply -f eso-logger.yaml
```

Follow the application logs:

```bash
kubectl logs -n eso-test deploy/eso-logger -f
```

Expected output:

```text
DB_PASSWORD=super-local-password
```

---

## 🐳 Docker Desktop Note

Some Docker Desktop Kubernetes installations cannot access locally built images and may return:

```text
ErrImageNeverPull
```

If this happens, push the image to the temporary **ttl.sh** registry:

```bash
IMAGE=ttl.sh/eso-logger-$RANDOM:1h

docker tag eso-logger:local $IMAGE
docker push $IMAGE

kubectl set image deployment/eso-logger eso-logger=$IMAGE -n eso-test

kubectl patch deployment eso-logger -n eso-test \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"eso-logger","imagePullPolicy":"Always"}]}}}}'
```

Verify the rollout:

```bash
kubectl rollout status deployment/eso-logger -n eso-test

kubectl logs -n eso-test deploy/eso-logger -f
```

---

## 🔄 Update the Password

Modify the value inside `eso-test.yaml`:

```yaml
value: super-local-password-changed
```

Apply the changes:

```bash
kubectl apply -f eso-test.yaml
```

Verify that ESO updated the Kubernetes Secret:

```bash
kubectl get secret demo-k8s-secret \
  -n eso-test \
  -o jsonpath='{.data.DB_PASSWORD}' | base64 -d
```

Expected output:

```text
super-local-password-changed
```

Continue watching the application logs:

```bash
kubectl logs -n eso-test deploy/eso-logger -f
```

After a short delay, the application automatically detects the new password:

```text
DB_PASSWORD=super-local-password
DB_PASSWORD=super-local-password
DB_PASSWORD=super-local-password-changed
```

---

## ⏱️ Why Isn't the Update Immediate?

The value of:

```yaml
refreshInterval: 2s
```

only controls **how frequently External Secrets Operator synchronizes the external provider into the Kubernetes Secret**.

Once the Kubernetes Secret is updated, the **kubelet** asynchronously refreshes the mounted Secret volume inside the Pod.

Finally, the application reads the updated file during its next polling cycle.

Because of these three independent steps, the application may observe the new value **a few seconds after** the Kubernetes Secret has already been updated.

---

## 💡 Best Practice

✅ Read secrets from **mounted Secret volumes** if your application needs to support secret rotation without restarting.

❌ Avoid using **environment variables** for rotating secrets, as they are only populated when the container starts and require a Pod restart to receive updated values.

---

## 🧹 Cleanup

Remove the application:

```bash
kubectl delete -f eso-logger.yaml --ignore-not-found
```

Remove the External Secret resources:

```bash
kubectl delete -f eso-test.yaml --ignore-not-found
```

Delete the test namespace:

```bash
kubectl delete namespace eso-test --ignore-not-found
```

Uninstall External Secrets Operator:

```bash
helm uninstall external-secrets -n external-secrets

kubectl delete namespace external-secrets --ignore-not-found
```

(Optional) Remove the local Docker image:

```bash
docker rmi eso-logger:local
```