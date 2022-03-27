# Kubernetes Setup Instructions


## Step 1: Install docker runtime 
By default, Kubernetes uses the Container Runtime Interface (CRI) to interface with your chosen container runtime.

If you don't specify a runtime, kubeadm automatically tries to detect an installed container runtime by scanning through a list of well known Unix domain sockets. The following table lists container runtimes and their associated socket paths:

**Runtime	Path to Unix domain socket**

Docker	/var/run/dockershim.sock

containerd	/run/containerd/containerd.sock

CRI-O	/var/run/crio/crio.sock

If both Docker and containerd are detected, Docker takes precedence. This is needed because Docker 18.09 ships with containerd and both are detectable even if you only installed Docker. If any other two or more runtimes are detected, kubeadm exits with an error.

See [container runtimes](https://kubernetes.io/docs/setup/production-environment/container-runtimes/) for more information.

Let's now continue with the installation:
```bash 
# remove older versions
sudo apt-get remove docker docker-engine docker.io containerd runc
# set up the repository
sudo apt-get update
sudo apt-get install \
    ca-certificates \
    curl \
    gnupg \
    lsb-release
# add docker's official gpg key
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
# set up the stable repository
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
# install docker engine
sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io

```

## Step 2: Install kubectl, kubeadm and kubelet
```bash 
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates curl
sudo curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
```


## Step 3: Configure the cgroup driver

Add the following line in /etc/systemd/system/kubelet.service.d/10-kubeadm.conf among the rest environment variables:
```
Environment="KUBELET_EXTRA_ARGS=--cgroup-driver=cgroupfs"
```
you can also use the following commands which will do it for you but I highly recommend that you do it on your own since there will probably be issues if the format of the file is changed:
``` bash
START=$(sudo head -n 2 /etc/systemd/system/kubelet.service.d/10-kubeadm.conf) 
END=$(sudo tail -n $(expr $(sudo wc -l /etc/systemd/system/kubelet.service.d/10-kubeadm.conf | cut -c 1-2) - 2) /etc/systemd/system/kubelet.service.d/10-kubeadm.conf) 
sudo truncate -s 0 /etc/systemd/system/kubelet.service.d/10-kubeadm.conf && sudo cat <<EOF | sudo tee -a /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
$START
Environment="KUBELET_EXTRA_ARGS=--cgroup-driver=cgroupfs"
$END
EOF

# verify that extra args environment variable has been added in the third line of your conf file:
sudo cat /etc/systemd/system/kubelet.service.d/10-kubeadm.conf
```
cgroupfs is the appropriate cgroup-driver for docker which we will be using as our container runtime for the sake of simplicity. If you want to use a more sophisticated container runtime for further sandboxing you might have to change this configuration.


## Step 4: Initialize the kube cluster:

```bash
# first disable paging and swapping since if memory swapping is allowed this can lead to stability issues when the scheduler tries to deploy a pod:
sudo swapoff -a
# delete previous plane if it exists:
sudo kubeadm reset
# remove previous configurations:
sudo rm -r ~/.kube
# initialize new control plane:
sudo kubeadm init --pod-network-cidr=10.244.0.0/16
# A couple of notes:
# 1. --pod-network-cidr=10.244.0.0/16 is the appropriate network for flannel cni which we will be using, --pod-network-cidr=192.168.0.0/16 is for calico
# 2. We could have also used --apiserver-advertise-address=<my_addr> to specify which ip we want the control plane to advertise to others, but since we didn't, kubelet will find the default network inteface and use its ip.
# 3. We could have also specified --cri-socket and use another cri socket (e.g. containerd.sock) in order to make kubernetes play with other container runtimes (such as gVisor) too but for now we will just keep things simple.  

# Move the appropriate configuration files in a place where kubernetes can find them, and give them the appropriate privileges:
mkdir -p $HOME/.kube
sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
sudo chown $(id -u):$(id -g) $HOME/.kube/config
# Alternatively if you are the root user: $ echo "export KUBECONFIG=/etc/kubernetes/admin.conf" | tee -a ~/.bashrc && source ~/.bashrc
# save the "kubeadm join" command printed in the last line of "kubeadm init" output in a file because you will need it to add workers in the cluster. 
```
## Step5: Apply the CNI 
**Flannel:**
```bash
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml
# Alternatively for calico:
kubectl create -f https://projectcalico.docs.tigera.io/manifests/tigera-operator.yaml
kubectl create -f https://projectcalico.docs.tigera.io/manifests/custom-resources.yaml
```

## Adding new nodes in your cluster:
In order to join more nodes in the cluster you have to repeat the first 3 steps on the new node and then run the join command you saved before as a root user in the new node

## A couple useful aliases I like to use for kubectl:
```bash
alias kgp="kubectl get pods --all-namespaces"
alias kgn="kubectl get nodes"
alias kgd="kubectl get deploy"
```

**You can also permanently add all of them in your shell script configuration with the following command:**
```bash
cat <<EOF | tee -a ~/.bashrc && source ~/.bashrc

alias kgp="kubectl get pods --all-namespaces"
alias kgn="kubectl get nodes --all-namespaces"
alias kgd="kubectl get deploy --all-namespaces"
EOF

# In case you are using zsh as the default shell for your user you could use the same command as above by replacing "bashrc" either with "zshrc" or "profile"
```

## From now on you can follow the main tutorial (k3d tutorial) in order to install prometheus and grafana using arkade, however in the following guide we will also show how to install prometheus and grafana using helm: 

Install helm which we will use for installing prometheus and grafana
``` bash
# install helm
curl https://baltocdn.com/helm/signing.asc | sudo apt-key add -
sudo apt-get install apt-transport-https --yes
echo "deb https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list
sudo apt-get update
sudo apt-get install helm
```
```bash
# install prom and grafana
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts

helm repo add stable https://charts.helm.sh/stable

helm repo update

helm install prometheus prometheus-community/kube-prometheus-stack

kubectl port-forward deployment/prometheus-grafana 3000
```
Verify the installation by visiting http:localhost:3000 and then the rest are exactly the same as in the main doc

**Important note**: If you are running in a real cloud cluster (e.g. gke, azure etc) then you will probably not be able to use the local registry we create in the main tutorial using docker for the test-db function and you will probably have to create docker-hub account and use a generic docker-hub registry in order to push images by replacing "registry.localhost:5000" in functions.test-db.image in [this file](https://github.com/dimgatz98/openfaas_plhroforiaka/tree/main/test-db-function/test-db.yml) with your appropriate registry. Furthermore, it is possible that you have issues with the cpu limits used in mongodb stateful set so you may have to change spec.template.spec.containers.resources.limits.cpu and spec.template.spec.containers.resources.requests.cpu to a fraction like "0.1" in [here](https://github.com/dimgatz98/openfaas_plhroforiaka/tree/main/mongodb/mongodb-stateful-deployment.yml).
