*This tutorial has only been tested on ubuntu 18.04 and 20.04*

# Deploying openfaas on a local k3d cluster

### First of all, clone the repo, open an ubuntu terminal and change your working directory withing the repo.

## Prerequisites

## 1. Install Docker Engine and kubernetes
Docker Engine is an open source containerization technology for building and containerizing your applications. 

Update the apt package index, and install the latest version of Docker Engine and containerd:
``` bash
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

Install kubelet, kubeadm and kubectl:
``` bash
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates curl
sudo curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update
sudo apt-get install -y kubelet kubeadm kubectl
sudo apt-mark hold kubelet kubeadm kubectl
```

## 2. k3d:
k3d is a lightweight wrapper to run k3s (Rancher Lab’s minimal Kubernetes distribution) in docker.
k3d makes it very easy to create single- and multi-node k3s clusters in docker, e.g. for local development on Kubernetes.

To install it run the following commands:

``` bash
wget -q -O - https://raw.githubusercontent.com/rancher/k3d/main/install.sh | bash
```

``` bash
# Let's create the volume we are going to use as the registry for openfaas to push and pull images among the pods
sudo docker volume create local_registry
```
``` bash
# Let's start the containerized registry
sudo docker container run -d --name registry.localhost -v local_registry:/var/lib/registry --restart always -p 5000:5000 registry:2
```
### Create k3d cluster with 3 agents

``` bash
cd k3d_registries
sudo k3d cluster create --volume $PWD/k3d-registries.yml:/etc/rancher/k3s/registries.yaml --agents 3
cd ../
``` 

``` bash
sudo docker network connect k3d-k3s-default registry.localhost
```

Move cluster configuration file to user so you can use it without sudo
```bash
mkdir ~/.kube
# move config to user 
sudo cp /root/.kube/config ~/.kube/config
# change permission to make config accessible by kubectl
chmod 777 ~/.kube/config
```

## 3. arkade
arkade provides a portable marketplace for downloading your favourite devops CLIs and installing helm charts, with a single command.
To install arkade run:

``` bash
curl -sLS https://get.arkade.dev | sudo sh
```

## 4.Openfaas

To install openfaas using arkade, run the following:
``` bash
# Install openfaas for synchronous requests
arkade install openfaas
# In the case that you want to use queue workers and make asynchronous requests to openfaas functions you have to use this command instead
arkade install openfaas --max-inflight=2
# For more info on async requests check this:
# https://docs.openfaas.com/reference/async/
# Install faas-cli
arkade get faas-cli
# and now move it to the /usr/local/bin/ folder for terminal to find it
sudo mv ~/.arkade/bin/faas-cli /usr/local/bin/
```
### Now that we have all the Prerequisites we can start with the deployments

# Deploy mongodb stateful set and create replica set within it:

``` bash
cd mongodb/
```
``` bash
kubectl apply -f .
```
``` bash
cd ../
```

Wait untill all three pods are deployed and then run:
``` bash
kubectl exec -it mongodb-replica-0 -n default -- mongo
```
### And within the mongo shell:
```
rs.initiate()
```
```
var cfg = rs.conf()
```
```
cfg.members[0].host="mongodb-replica-0.mongo.default.svc.cluster.local:27017"
```
```
rs.reconfig(cfg)
```
```
rs.add("mongodb-replica-1.mongo.default.svc.cluster.local:27017")
```
```
rs.add("mongodb-replica-2.mongo.default.svc.cluster.local:27017")
```

### check if everything applied:
```
rs.status()
```
### create root user:
```
use admin
```
for the sake of simplicity we use a naive password "123", feel free to replace the password with your own
```
db.createUser({user: "root",pwd: "123",roles: [ "root" ]})
```
```
exit
```

### Now port-fortward openfaas' service and login via your terminal:
``` bash
# port forward
kubectl port-forward -n openfaas svc/gateway 8080:8080
# save password to $PASSWORD env variable
PASSWORD=$(kubectl get secret -n openfaas basic-auth -o jsonpath="{.data.basic-auth-password}" | base64 --decode; echo)
# login using the variable
echo -n $PASSWORD | sudo faas-cli login --username admin --password-stdin
```

## Create openfaas secret to store root password:
``` bash
# In a new terminal
export MONGODB_ROOT_PASSWORD="123"
```
## port forward openfaas and login from terminal:
``` bash
faas-cli secret create mongo-db-password --from-literal $MONGODB_ROOT_PASSWORD
```
### And now we are ready to deploy test-db, also with 3 replica as in mongo:
### The function we made is a simple function that leverages mongodb persistent storage to save the usernames of people we want to follow on social media. POST/PUT adds a new record in the db(no duplicates allowed), GET returns the list of people we have saved and DELETE deletes a user if he exists(you can also provide "all" in request body data in order to drop the whole collection)
``` bash
cd test-db-function/
# If you want to use asynchronous requests you have to define max-inflight as an environment variable
# synchronous example
sudo faas-cli up -f test-db.yml --label com.openfaas.scale.max=3 --label com.openfaas.scale.min=3
# asynchronous example
sudo faas-cli up -f test-db.yml --label com.openfaas.scale.max=2 --label com.openfaas.scale.min=2 --env max_inflight=1000
cd ../
# You can also always check the "queue-worker"'s logs like that
kubectl logs deploy/queue-worker -n openfaas -f
```

Now you are ready to hit the endpoint either via browser (by visiting http://localhost:8080/) and making manual requests or via the terminal by executing either one of the following commands:
``` bash
# For adding user with username "random_user" in mongodb:
curl -X POST -d "random_user" http://localhost:8080/function/test-db
# or 
curl -X PUT -d "random_user" http://localhost:8080/function/test-db
# For querying all data in database:
curl -X GET http://localhost:8080/function/test-db
# For deleting user with username "random_user"
curl -X DELETE -d "random_user" http://localhost:8080/function/test-db
# Check the keys again:
curl -X GET http://localhost:8080/function/test-db
# Output should be empty after deletion
```

# Metrics

In order to export our metrics we will be using Prometheus along with Grafana in order to graph them.

Prometheus is free and an open-source event monitoring tool for containers or microservices. Prometheus collects numerical data based on time series. The Prometheus server works on the principle of scraping. This invokes the metric endpoint of the various nodes that have been configured to monitor. These metrics are collected in regular timestamps and stored locally. The endpoint that was used to discard is exposed on the node.

Grafana is a multi-platform visualization software available since 2014. Grafana provides us a graph, the chart for a web-connected to the data source. It can query or visualize your data source, it doesn’t matter where they are stored.

## Prerequisites


## 1. Grafana
You can install Grafane via arkade using:

``` bash
arkade install grafana
```
### Get secret password:
``` bash
kubectl get secret --namespace grafana grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo
```
Save this secret password as you will need it for logging into the Grafana dashboard.

# Apache JMeter
The Apache JMeter™ application is open source software, a 100% pure Java application designed to load test functional behavior and measure performance. It was originally designed for testing Web Applications but has since expanded to other test functions.

## Step 1: Install Java 

Apache JMeter is a Java-base application so Java must be installed in your system. You can install it by running the following command:

``` bash
sudo apt-get install openjdk-8-jdk -y
```

## Step 2: Install Apache Web Server 

You can install it with the following command:
``` bash
sudo apt-get install apache2 -y
```

After installing Apache web server, start the Apache service and enable it to start at system reboot:
``` bash
sudo systemctl start apache2
```
``` bash
sudo systemctl enable apache2
```

## Step 3: Install Apache JMeter 

You can download it with the following command:
``` bash
wget https://downloads.apache.org//jmeter/binaries/apache-jmeter-5.3.zip
```
Once downloaded, unzip the downloaded file with the following command:
```
unzip apache-jmeter-5.3.zip
```

## Step 4: Launch the Apache JMeter Application

Next, change the directory to the JMeter:

``` bash
cd apache-jmeter-5.3/bin
```

Now, start the JMeter application with the following command:

``` bash
./jmeter
```
You should see the JMeter interface. From there you can generate your load in order to monitor your system.
**Note:** In jmeter requests body data you can provide the option "${__UUID()}" in order to require jmeter to add a different UUID in each request's data since our database makes sure that the post request's data(username) is different than all existing followers currently in the database before adding the new record.  

We also created our own load generator [here](https://github.com/dimgatz98/openfaas_plhroforiaka/tree/main/load_generator) and a command line interface for extracting and visualising metrics [here](https://github.com/dimgatz98/openfaas_plhroforiaka/tree/main/prom_metrics_cli).

### To monitor the health and behavior of our function we have to:

## 1. Port fortward Prometheus
```
kubectl port-forward -n openfaas svc/prometheus 9090:9090
```

## 2. Port forward Grafana

```
kubectl --namespace grafana port-forward service/grafana 3000:80
```

Finally, you can visit prom and grafana dashboards here:
```
http:/localhost:9090/
```
and here:
```
http:/localhost:3000/
```
respectively and start monitoring your application. 

**Note**  that in order to be able to graph your data through grafana you have to first add prometheus data source from:
Configuration (on the left bar) > Data Sources > Add data source > Select Prometheus > Fill URL with "http:/localhost:9090", in the "Access field" choose "Browser" and then hit "Save & Test".

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

### In order to remove the whole project you have to run the following commands:
```bash 
# remove openfaas
kubectl delete namespace openfaas openfaas-fn
# remove mongodb
cd mongodb/
kubectl delete -f .
cd ..
```
### If you want to completely delete the k3d cluster you have to run the following command:
```bash
sudo k3d cluster delete <cluster_name>
# where cluster name can be found from the following command
sudo k3d cluster list 
```

If you want to learn how to setup kubernetes using kubeadm click [here](https://github.com/dimgatz98/openfaas_plhroforiaka/tree/main/kubeadm_tutorial/README.md).

