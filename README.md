# Gazo-san-controller
Gazo-san-controller is a custom controler for the Gazo-san CRD on Kubernetes.

Based on the [Gazo-san-report](https://github.com/rueyaa332266/gazo-san-report), creating the image difference report by passing the URLs in the CRD.

## Gazo-san CRD sample
```yaml
apiVersion: gazosancontroller.k8s.io/v1alpha1
kind: Report
metadata:
  name: report-sample
spec:
  baseURL: "URL for base site"
  compareURL: "URL for compare site"
```

## Setup

> Note: Make sure you install kustomize first❗

1. Clone the source code.
```
git clone git@github.com:rueyaa332266/gazo-san-controller.git
cd gazo-san-controller
```

2. Install Gazo-san CRD into the cluster.
```
make install
```

3. Deploy Gazo-san-controller into the cluster.
```
export IMG=aa332266/gazo-san-controller:kubebuilder
make deploy
```

## Usage
After deploying Gazo-san-controller in your cluster, you can use the Gazo-san CRD normally.

1. Create a Gazo-san CRD manifest. (You can just copy the Gazo-san CRD sample in README.)

2. Set the URL you want to compare.
```yaml
apiVersion: gazosancontroller.k8s.io/v1alpha1
kind: Report
metadata:
  name: report-sample
spec:
  baseURL: https://www.google.com/?hl=en
  compareURL: https://www.google.com/?hl=ja
```

3. Apply the manifest.
```
kubectl apply -f gazo-san-report.yml
```

Gazo-san-controller will create a pod for your report page, using nginx port 80.

4. Checking the report

The simple way is port forwarding the pod.
```
kubectl get pod
kubectl port-forward pod/gazo-san-report 8080:80
```

## Remove Gazo-san-controller
```
kustomize build config/default | kubectl apply -f -
kustomize build config/crd | kubectl delete -f -
```

## Build by
[Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder)