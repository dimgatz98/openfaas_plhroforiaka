# Deploying openfaas on a local k3d cluster

## Prerequisites

## 1. Docker Engine
Docker Engine is an open source containerization technology for building and containerizing your applications. 

Update the apt package index, and install the latest version of Docker Engine and containerd:
```
 sudo apt-get update
 ```
 ```
 sudo apt-get install docker-ce docker-ce-cli containerd.io
```

## 2. k3d:
k3d is a lightweight wrapper to run k3s (Rancher Lab’s minimal Kubernetes distribution) in docker.
k3d makes it very easy to create single- and multi-node k3s clusters in docker, e.g. for local development on Kubernetes.

To install it run the following commands:

```
wget -q -O - https://raw.githubusercontent.com/rancher/k3d/main/install.sh | bash
```
```
sudo docker volume create local_registry
```
```
sudo docker container run -d --name registry.localhost -v local_registry:/var/lib/registry --restart always -p 5000:5000 registry:2
```

## 3. arkade
arkade provides a portable marketplace for downloading your favourite devops CLIs and installing helm charts, with a single command.
To install arkade run:

```
curl -sLS https://get.arkade.dev | sudo sh
```

## 4.Openfaas

TO install openfaas using arkade run:
```
arkade get faas-cli
```
### Now that we have all the Prerequisites we can create a k3d cluster with 3 agents and connect it to the network

# Create k3d cluster with 3 agents

```
sudo k3d cluster create --volume $PWD/k3d-registries.yml:/etc/rancher/k3s/registries.yaml --agents 3
```
```
sudo docker network connect k3d-k3s-default registry.localhost
```

# Deploy mongodb stateful set and create replica set within it:

```
cd mongodb/
```
```
sudo kubectl apply -f .
```

Wait untill all three pods are deployed

And then run:
```
sudo kubectl exec -it mongodb-replica-0 -n default -- mongo
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
```
db.createUser({user: "root",pwd: "123",roles: [ "root" ]})
```
```
exit
```

## Create openfaas secret to store root password:
```
export MONGODB_ROOT_PASSWORD="123"
```
## port forward openfaas and login from terminal:
```
faas-cli secret create mongo-db-password --from-literal $MONGODB_ROOT_PASSWORD
```
### And now we can deploy test-db:
```
sudo faas-cli up -f test-db.yml --label com.openfaas.scale.max=3 --label com.openfaas.scale.min=2
```

# Metrics

We can use prometheus and grafana.

Prometheus is free and an open-source event monitoring tool for containers or microservices. Prometheus collects numerical data based on time series. The Prometheus server works on the principle of scraping. This invokes the metric endpoint of the various nodes that have been configured to monitor. These metrics are collected in regular timestamps and stored locally. The endpoint that was used to discard is exposed on the node.

Grafana is a multi-platform visualization software available since 2014. Grafana provides us a graph, the chart for a web-connected to the data source. It can query or visualize your data source, it doesn’t matter where they are stored.

## Prerequisites


## 1. Grafana
You can install Grafane via arkade using:

```
sudo arkade install grafana
```
### Get secret password:
```
sudo kubectl get secret --namespace grafana grafana -o jsonpath="{.data.admin-password}" | base64 --decode ; echo
```
Save this secret password. You will need it.

# Apache JMeter
The Apache JMeter™ application is open source software, a 100% pure Java application designed to load test functional behavior and measure performance. It was originally designed for testing Web Applications but has since expanded to other test functions.

## Step 1: Install Java 

Apache JMeter is a Java-base application so Java must be installed in your system. You can install it by running the following command:

```
apt-get install openjdk-8-jdk -y
```

## Step 2: Install Apache Web Server 

You can install it with the following command:
```
apt-get install apache2 -y
```

After installing Apache web server, start the Apache service and enable it to start at system reboot:
```
systemctl start apache2
```
```
systemctl enable apache2
```

## Step 3: Install Apache JMeter 

You can download it with the following command:
```
wget https://downloads.apache.org//jmeter/binaries/apache-jmeter-5.3.zip
```
Once downloaded, unzip the downloaded file with the following command:
```
unzip apache-jmeter-5.3.zip
```

## Step 4: Step 5 - Launch the Apache JMeter Application

Next, change the directory to the JMeter:

```
cd apache-jmeter-5.3/bin
```

Now, start the JMeter application with the following command:

```
./jmeter
```
You should see the JMeter interface.

### To monitor the health and behavior of our function we can run:

## 1. Run prometheus
```
sudo kubectl port-forward -n openfaas svc/prometheus 9090:9090
```

## 2. Run Grafana

```
sudo kubectl --namespace grafana port-forward service/grafana 3000:80
```
