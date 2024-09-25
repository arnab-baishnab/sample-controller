# My Kluster Controller

This project provides a custom Kubernetes controller for managing Kluster resources. It simplifies the deployment and management of applications within a Kubernetes cluster.

## Table of Contents
- [Installation](#installation)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)
- [Contact](#contact)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/arnab-baishnab/sample-controller.git

#Navigate to the project directory:

```bash
cd sample-controller
Install dependencies:
```

```bash
go mod tidy
Build the project:
```

#go build
##Install the Custom Resource Definition (CRD):

```bash
kubectl apply -f manifests/arnabbaishnab.com/Updated_group_name_klusters.yaml
```
#Usage
##To create a new Kluster resource, use the following command:

```bash
kubectl apply -f manifests/arnabbaishnab.com/example1.yaml
```
