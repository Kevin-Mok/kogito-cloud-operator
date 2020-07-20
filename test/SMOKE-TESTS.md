# Guide for Core/Runtimes Team to Smoke Test Local Changes on OpenShift/K8s Cluster

Before starting to build the artifacts with your changes, 
you will need to set up a Nexus instance locally. We chose to deploy 
it on a K8s cluster as we’ll be using that for testing the 
operator.

## Table of Contents
* [Install Minikube](#install-minikube)
* [Setup Nexus](#setup-nexus)
   * [Getting Admin Password](#getting-admin-password)
* [Setup Maven Repository](#setup-maven-repository)
* [Build and Deploy Artifacts](#build-and-deploy-artifacts)
* [Building the Images](#building-the-images)
   * [Packages](#packages)
   * [Python Modules](#python-modules)
   * [Updating Maven Information](#updating-maven-information)
   * [Building Images](#building-images)
      * [Image List](#image-list)
* [Installing the Operator](#installing-the-operator)

Created by [gh-md-toc](https://github.com/ekalinin/github-markdown-toc).

## Install Minikube
We decided to go with Minikube as the K8s cluster as it is 
very resource efficient and can be started easily on any 
system. 

**Prerequisites**:  
* [Install kubectl binaries](https://kubernetes.io/docs/tasks/tools/install-kubectl/) on your system 
* Have a hypervisor installed (kvm is recommended for Linux)

For installing the Minikube cluster, please follow 
[this tutorial](https://kubernetes.io/docs/tasks/tools/install-minikube/).

## Setup Nexus
We are going to use the Nexus operator from 
[m88i labs](https://github.com/m88i/nexus-operator/) to deploy the Nexus.

Follow the steps to have your Nexus:
* Clone the above repo.
* Run `make install` in the cloned repo.
* This will expose your Nexus server as a `NodePort` service on Minikube. This 
  will be accessed from the Minikube's IP (`minikube ip` will return the IP address of your Minikube cluster).
* To find the port on which Nexus server is running on the above IP, run:
    ```
    $ kubectl get svc -n nexus
    NAME                     TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
    nexus-operator-metrics   ClusterIP   10.103.2.155    <none>        8383/TCP,8686/TCP   30m
    nexus3                   NodePort    10.109.238.40   <none>        8081:31031/TCP      30m
    ```
* You can just run `minikube service nexus3 -n nexus` to 
  open the Nexus server on your default browser.
  * *Note*: It can take a few minutes to have the Nexus server running.

### Getting Admin Password
For authentication with the Nexus server, you’ll need to get 
the admin password from the pod running the Nexus server. 
Check the pod name:
```
$ kubectl get pod -n nexus
NAME                              READY   STATUS    RESTARTS   AGE
nexus-operator-8577d97489-57w8n   1/1     Running   0          35m
nexus3-66f6555ff-zqc59            1/1     Running   0          35m
```

The `nexus3-*` is your pod with Nexus server running. The 
initial admin password is stored at 
`/nexus-data/admin.password` on the server. 

You can get that by running the following command:
`kubectl exec nexus3-66f6555ff-zqc59 -n nexus -- cat 
/nexus-data/admin.password` (just remember to change the pod 
name from the one displayed in your setup).

You can use this password and the username `admin` to 
authenticate with your Nexus. It’ll ask to change it in the 
initial setup. Make sure to allow for anonymous access to 
the repository.

## Setup Maven Repository

Next we have to set up a hosted Maven repository for storing 
the artifacts building from runtimes/apps/examples. 

After initial setup and allowing anonymous users access to read, 
do the following:
1. Go to the gear icon in the top-left corner and select `Repositories` from the left panel.
1. Select `Create Repository` and `maven2 (hosted)` from the list.
1. Give it a name (e.g. `runtimes-artifacts`).
1. Make sure `Online` is checked and select `Mixed` in the `Version Policy -> What type of artifacts does this repository store?`.
1. Select `Allow redeploy` (in case of any errors) inside `Hosted -> Deployment Policy`.
1. You can leave everything else as default and click on `Create Repository`. The repository URL can be seen after creating the repository from its home page.
1. Click on the gear icon again and select `Security -> Users` from the left sidebar. Select `anonymous` user, move 
`nx-admin` from the available roles to the granted roles and 
click save. 
    - This is to allow anonymous users to push to this 
repository as this repository is hosted inside Minikube in 
your system; it won't be accessible by anyone else, so we 
need to enable anonymous push to this Nexus server. 

The repository is all ready now; we can build and deploy our artifacts in it.

## Build and Deploy Artifacts
Before proceeding, read the `README` files on the following repos to know the dependencies you might need to install to build and deploy the artifacts:
- [kogito-runtimes](https://github.com/kiegroup/kogito-runtimes#requirements)
- [kogito-apps](https://github.com/kiegroup/kogito-apps#building-from-source)
- [kogito-examples](https://github.com/kiegroup/kogito-examples#process-hello-world-with-scripts)

We will deploy the kogito-runtimes artifacts first. Copy 
your repository URL from Nexus by clicking on the gear icon, 
selecting `Repository -> Repositories` from the sidebar and 
clicking the copy button in your repository's row. Then, 
run the following commmand inside your cloned kogito-runtimes repo:
```
mvn clean deploy -DaltDeploymentRepository=repository_name_here::default::repository_url_here
```
This command will build, test and deploy your artifacts on your repository. If you want to skip tests, run:
```
mvn clean deploy -DaltDeploymentRepository=repository_name_here::default::repository_url_here -DskipTests
```

The same command can be used to deploy the apps/examples 
artifacts inside their respective repositories.

## Building the Images
After all the artifacts are present in the Maven repository, 
we start to build the 
[kogito-images](https://github.com/kiegroup/kogito-images). 
We use CEKit to build our images, so some dependencies are 
needed.

### Packages 
* `docker`
* `gcc`
* `krb5-devel`
* `s2i`
* `python3`
* GraalVM 19.3.1 (Java 11 or higher)
* `native-image`
  - installed by running `gu install native-image`
  - `gu` is present after Graal is installed and in `PATH`
* `zlib-devel`
* `glibc-devel`
* OpenJDK 11.0.6 or higher
* Maven 3.6.2 or higher
* `pip3` 
  - Can be installed from [here](https://pip.pypa.io/en/stable/installing/). Just keep in mind to run the script with `python3`.

### Python Modules
Run: 
```
pip3 install cekit behave lxml docker docker-squash odcs elementpath pyyaml ruamel.yaml
```

### Updating Maven Information
After downloading the packages, you will need to update the 
Maven information for the images. You can use the 
script inside the kogito-images repository. Inside the cloned repository, run: 
```
python3 scripts/update-maven-information.py --repo-url=repository_url_here
```

### Building Images

For building all the images, you can run `make ignore_test=true` (if you want 
to skip image tests). 

If you only want to build and test individual images, you would first need to run `make clone-repos`.
Then for individual images, you can follow the syntax: `make 
<image_name> ignore_test=true` (optional, set `true` for 
skipping the tests).

#### Image List
* kogito-quarkus-ubi8
* kogito-quarkus-jvm-ubi8
* kogito-quarkus-ubi8-s2i
* kogito-springboot-ubi8
* kogito-springboot-ubi8-s2i
* kogito-data-index
* kogito-jobs-service
* kogito-management-console

## Installing the Operator

1. Enable the OLM in our Minikube cluster: `minikube addons enable olm`.
2. Clone the [operator-lifecycle-manager](https://github.com/operator-framework/operator-lifecycle-manager) repo.
3. From the root of the project, run `make 
run-console-local`. This will run the OperatorHub console on 
http://localhost:9000. 
    - You will need to have `jq` installed 
and port 9000 available on the system.
4. Create a different namespace where kogito-operator and all the dependent operators will run: `kubectl create ns kogito`.
5. On your browser, visit https://localhost:9000. Select `Operators > OperatorHub` and search for "kogito". 
Select the Kogito Operator by Red Hat. Install it with the default 
options, only changing the namespace where it needs to be installed. You can select the `kogito` namespace which was created for this purpose.
6. You can see the pods by running `kubectl get pod -n kogito`. Wait until all pods are in running state. It is installing kogito-operator and all the dependent operators for Kogito.
