Guide for Core/Runtimes team to Smoke test local changes on Openshift/k8s Cluster.


Before starting to build the artifacts with your changes. You would need to set up a nexus locally, we chose to deploy it on a k8s cluster as we’ll be using that for testing the operator.


Install Minikube: 


We decided to go with Minikube as the k8s cluster as it is very resource efficient and can be started easily on any system. 


Prerequisites:  
* Install kubectl binaries on your system [see](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 
* Have a hypervisor installed (kvm  is recommended for linux)


Finally for installing the Minikube cluster please follow [this tutorial](https://kubernetes.io/docs/tasks/tools/install-minikube/)


Setup Nexus:[a]
We are going to use the nexus operator from 
[m88i labs](https://github.com/m88i/nexus-operator/) to deploy the nexus.


Follow the steps to have your nexus:
* Clone the above repo.
* Run `make install` in the cloned repo.
* This will expose your nexus server as a nodePort service on Minikube. Which can be accessed from the Minikube’s IP
* `# minikube ip` (Will return the IP address of your minikube cluster.)
To find the port on which nexus server is running on the above IP run:
 ```
kubectl get svc -n nexus
NAME                         TYPE            CLUSTER-IP          EXTERNAL-IP   PORT(S)                 AGE
nexus-operator-metrics   ClusterIP   10.107.88.110   <none>            8383/TCP,8686/TCP   17m
nexus3                       NodePort        10.110.72.0         <none>            8081:31031/TCP          17m
```
You can just run 
`# minikube service nexus3 -n nexus`
It’ll open the nexus server on your default browser


Note: It can take a few minutes to have the nexus server running


For authentication with the nexus server, you’ll need to get the admin password from the pod running the nexus server.[b]


Check the pod name.
```
kubectl get po -n nexus
NAME                                          READY       STATUS            RESTARTS   AGE
nexus-operator-8577d97489-zr4wg   1/1                 Running         0              15m
nexus3-66f6555ff-pkbjx                        1/1                 Running         0              14m
```
The with nexus3-* is your pod with nexus server running, the initial admin password is stored at `/nexus-data/admin.password` on the server.


You can get that by running the following command.
# `kubectl exec nexus3-66f6555ff-pkbjx -n nexus -- cat /nexus-data/admin.password`
 (Just remember to change the pod name from the one displayed in your setup)


You can use this password and admin username to authenticate with your nexus, it’ll ask to change it in the initial setup. Make sure to allow for anonymous access to the repository.


Next we’ve to set up a hosted maven repository for storing the artifacts building from runtimes/ apps/ examples. 




Setup Maven Repository:


After Initial setup and allowing anonymous users access to read. 
* Go to the gear icon in the left corner and select repositories from the left panel
* Select `Create Repository` and then select `maven2 (hosted)` from the list
* Give it a name for eg: `runtimes-artifacts`.
* Make sure you check the repository is online and select `mixed` in the `Version Policy -> What type of artifacts does this repository store?`
* Also allow redeploy in case of any errors, inside `Hosted -> Deployment Policy`
* You can leave everything as default and click on `Create Repository` 
* The repository URL can be seen after creating the repository from its home page
* Allow anonymous users to push to this repository (as this repository is hosted inside minikube in your system it won’t be accessible by anyone else, so we can enable the anonymous push to this nexus server.)
* For that Click on the Gear icon -> Select Users (Inside security) -> Select Anonymous User -> Grant admin role to the user and click save.


The repository is all ready now, we can build and deploy our artifacts in it 




Build and Deploy Artifacts:[c]


Before proceeding read the README files on the following repos to know the dependencies you might need to install to build and deploy the artifacts
* Let’s deploy the [kogito-runtimes](https://github.com/kiegroup/kogito-runtimes) artifacts first. 
Inside your cloned kogito-runtimes repo run
```
mvn clean deploy -DaltDeploymentRepository=runtimes-artifacts::default::http://172.17.0.3:31031/repository/runtimes-artifacts/   -DskipTests(optional if you want to skip tests)
```
The above command will build, test and deploy your artifacts on your repository.


Note: The repository url will be different in your case, though you’ll be able to see it when you select your repository from the repositories list on the nexus server.


* Similarly the [kogito-apps](https://github.com/kiegroup/kogito-apps) artifacts can be deployed with the same command inside the cloned repository.
```
mvn clean deploy -DaltDeploymentRepository=runtimes-artifacts::default::http://172.17.0.3:31031/repository/runtimes-artifacts/   -DskipTests(optional if you want to skip tests)
```


* Same for the [kogito-examples](https://github.com/kiegroup/kogito-examples) artifacts can be deployed with the same command inside the cloned repository.
```
mvn clean deploy -DaltDeploymentRepository=runtimes-artifacts::default::http://172.17.0.3:31031/repository/runtimes-artifacts/   -DskipTests(optional if you want to skip tests) ```




Build the Images:[d]
After all the artifacts are present in the maven repository, we start to build the [kogito-images](https://github.com/kiegroup/kogito-images). We use a cekit to build our images so few dependencies are needed.


Prerequisites:


 
Packages: 
* docker
* gcc
* krb5-devel
* s2i
* python3
* GraalVM 19.3.1-java11 +
* native-image(installed by running gu install native-image, gu is present after graal is installed and in path)
* zlib-devel
* glibc-devel
* OpenJDK 11.0.6
* Maven 3.6.2+
* pip3 (Can also be installed from [here](https://pip.pypa.io/en/stable/installing/) Just keep in mind to run the script with python3)






        Python Modules: needs to be installed by pip3. By running pip3 install <package>
* cekit
* behave
* lxml
* docker
* docker-squash
* odcs[client]
* elementPath
* pyyaml
* ruamel.yaml












After downloading the packages you would need to update the maven information for the images. For that you can use the script inside the kogito-images repository.


Inside your cloned repository just run:
`python3 scripts/update-maven-information.py  --repo-url=http://172.17.0.3:31031/repository/runtimes-artifacts/`


Note: Please keep in mind to update the repo-url with the repository that you deployed previously and stored all your artifacts on.


Now let’s build the images.


For building all the images you can run `make ignore_test=true(if you want skip image tests)`


If you only want to build and test individual images you would first need to run:


# make clone-repos
And for individual images you can follow the syntax below        


# make <image_name> ignore_test=true(optional, set true for skipping the tests)


Image List:
* kogito-quarkus-ubi8
* kogito-quarkus-jvm-ubi8
* kogito-quarkus-ubi8-s2i
* kogito-springboot-ubi8
* kogito-springboot-ubi8-s2i
* kogito-data-index
* kogito-jobs-service
* kogito-management-console
























Install the Operator:


First we would need to enable the OLM in our Minikube cluster for that just run:


`$ minikube addons enable olm`


Launch the OLM console locally, 
Just clone the [operator-lifecycle-manager](https://github.com/operator-framework/operator-lifecycle-manager) repo and from the root of the project run: `$ make run-console-local`. This will run the operatorhub console on http://localhost:9000 
Note: Needs to have `jq[e]` installed and 9000 port available on the system.


Create a different namespace where kogito-operator and all the dependent operator will run


`$ kubectl create ns kogito`


Now on your browser and visit  https://localhost:9000. Select `Operators > OperatorHub` and search for kogito. 
Select the kogito operator by Redhat and install it with defaults only changing the namespace where it needs to be installed. You can select the `kogito` namespace which was created for this purpose.


You can see the pods by:


`$ kubectl get po -n kogito`
Wait till all pods are in running state (It is installing kogito-operator and all the dependent operator for kogito)
[a]I believe this step could be optional to improve the build time only.
[b]Oh we added the `admin/admin123` on 0.3.0.. we need to release that soon.
[c]Instead of building the applications, one can use our images and just use a custom Dockerfile to build on top of it. There's no need to build and deploy the runtimes entirely.
[d]Here, a developer could use Dockerfiles with our Kogito images instead. Would improve the time taken to build an app.
[e]Let's add the link with more info  about this package.
