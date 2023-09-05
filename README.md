## Volume-Snapshotter

The aim of this project is to create a kubernetes controller to snapshot a volume and also create a PVC out of an existing volume snapshot

### Implementation

In order to achieve the above task I have created my own CRD that can be found at pkg/apis/pranoykundu.dev/v1/types.go.
This project heavily uses the Volume Snapshot API and the PVC API.

### Setup

In order run the application we need to setup CSI driver as mentioned here https://github.com/kubernetes-csi/csi-driver-host-path/blob/master/docs/deploy-1.17-and-later.md
All the required files are mentioned in the manifests/