Here's how to configure Astra Proxy Service to capture the HTTP traffic for Postman/Burp/Curl. It also covers troubleshooting common issues encountered during configuration.

### Astra Proxy integration

#### Quick Installation

1. **SSH into the VM or developer box where proxy should be hosted.**

2. **Pre-requisite**

* Install Docker in your VM by following the official [doc](https://docs.docker.com/engine/install/)

3. ** Download *astra-cli* from [repository]()**

4. **Create an environment file called as *proxyEnv* and fill below contents**

* Replace < **sensorID**> with integration ID displayed for your astra-proxy-service integration [page](https://my.getastra.com/integrations) in UI.
* Replace < **allowed_hosts**> with the , separated list of host. Minimum one, maximum five entries can be given as comma separated host names
* Replace < **host:port**> with the address of the [astra-traffic-collector](https://help.getastra.com/en/article/how-to-setup-astra-traffic-collector-in-vm-1dile2q/?bust=1732789353824)

```
SENSOR_ID=<sensorID>
ALLOWED_HOSTS=<allowed_hosts>
OTEL_EXPORTER_ENDPOINT=<host:port>
```

5. ** Run the following to start astra-proxy-service**

* For linux:
```
chmod +x astra-cli
./astra-cli proxy quickstart --listen-port 8181 --env-file proxyEnv
```

* For Windows:
 
```
./astra-cli.exe proxy quickstart --listen-port 8181 --env-file proxyEnv
```

* For Mac:
```
./astra-cli proxy quickstart --listen-port 8181 --env-file proxyEnv
```

6. **To check the status of astra-proxy-service, exec below command**
* For linux:
```
./astra-cli proxy status
```

* For Windows:
 
```
./astra-cli.exe proxy status
```

* For Mac:
```
./astra-cli proxy status
```

7. **Set the proxy address of astra-proxy-service as upstream proxy in your application like [postman](https://learning.postman.com/docs/getting-started/installation/proxy/), [burp](https://portswigger.net/burp/documentation/desktop/settings/network/connections#upstream-proxy-servers) or [curl](https://www.zenrows.com/blog/curl-proxy#set-up-curl-proxy) and then start using your application**

#### Upgrade

##### Docker container upgrade

This process updates the docker container to a new version of the astra-proxy-service.

1. ** Change directory to the place where *astra-cli* executable is downloaded**

2. **Run below command**
```
./astra-cli proxy upgrade
```

3. **Upon successfull image pull, run this to stop the current container and subsequently remove it**
```
./astra-cli proxy stop
./astra-cli proxy remove
```

4. **Restart the container with newly pulled image**
```
./astra-cli proxy quickstart --listen-port 8181 --env-file proxyEnv
```








Astra-cli is a wrapper tool around docker to manage astra-proxy-service on the fly. It can be used to setup and manage the astra-proxy-service by launching this proxy service as a container. This guide will cover how to use astra-cli to manage astra-proxy-service. It also covers troubleshooting common issues encountered for astra-proxy-service.

|| astra-proxy-service makes use of well known [mitmproxy](https://mitmproxy.org/#mitmdump) as upstream proxy server. This service by default doesn't verify the upstream certificates and hence the certificate verification is left to the application

#### Download *astra-cli* from repository

1. **Open your browser and download the astra-cli from here for your preferred operating system.**

2. **To run the astra-cli in your system, following pre-requisite should be met:**

* Install Docker in your system by following the official [doc](https://docs.docker.com/engine/install/)

#### Manage *astra-proxy-service* by using astra-cli

1. **Environment file is mandatory for the astra-proxy-service to start**

* Create an env file called as **proxyEnv** and add **SENSOR_ID**, **ALLOWED_HOSTS**,  **OTEL_EXPORTER_ENDPOINT** entries to this env file where:
   
SENSOR_ID is the integrationID displayed in the [integrations page](https://my.getastra.com/integrations) of getastra   
ALLOWED_HOSTS is the comma separated list of host names. Minimum one, maximum five comma separated entries can be given.
OTEL_EXPORTER_ENDPOINT is the address of the [astra-traffic-collector](https://help.getastra.com/en/article/how-to-setup-astra-traffic-collector-in-vm-1dile2q/?bust=1732789353824)

* Example env file is shown below

```
SENSOR_ID=f0dd7367-5f66-4c1b-bd73-74da8a5b78a6
ALLOWED_HOSTS=mydomain.dev, mydomain.com, testing.com
OTEL_EXPORTER_ENDPOINT=localhost:4317
```

2. **Start the astra-proxy-service container under quickstart mode**

* quickstart mode accepts two parameters, **--listen-port** and **--env-file** where:

--listen-port : will set the port on which http proxy should be listening to
--env-file : will set the env file to read from

* Following command will start a simple http proxy server by binding the astra-proxy-service to host network. Proxy will be accessible at address http://localhost:8181
```
./astra-cli proxy quickstart --listen-port 8181 --env-file proxyEnv
```

3. **Start the astra-proxy-service with additional flags**

* astra-cli being a wrapper around docker, the cli supports almost all the flags supported by docker [run](https://docs.docker.com/engine/containers/run/#general-form). Additionally, the astra-proxy-service makes use of well known mitm proxy, and hence supports all the flags supported by [mitmdump](https://docs.mitmproxy.org/stable/concepts-options/#available-options)

* Following is a sample command which will start astra-proxy-service with docker container port mapping.
```
./astra-cli proxy start --env-file .env --rm -p 8080:8181
```

4. **Check the status of astra-proxy-service**

```
./astra-cli proxy status
```
* You should see similar output like this

```
CONTAINER ID   IMAGE            COMMAND                  CREATED          STATUS          PORTS                                                 NAMES
4e0090bb1ae9   getastra/proxy   "mitmdump -k -s /app…"   35 minutes ago   Up 35 minutes   8080/tcp, 0.0.0.0:8080->8181/tcp, :::8080->8181/tcp   astra-proxy-service
```

5. **Check the logs of astra-proxy-service**

```
./astra-cli proxy logs
```

* To tail the logs:
```
./astra-cli proxy logs --tail=0 -f
```

* To check the logs for last 5 minutes
```
./astra-cli proxy logs --since=5m
```

6. **To stop astra-proxy-service**

```
./astra-cli proxy stop
```

7. **To stop astra-proxy-service**

```
./astra-cli proxy remove
```

#### Upgrade

##### Docker container upgrade

This process updates the docker container to a new version of the astra-proxy-service.

1. ** Change directory to the place where *astra-cli* executable is downloaded**

2. **Run below command**
```
./astra-cli proxy upgrade
```

3. **Upon successfull image pull, run this to stop the current container and subsequently remove it**
```
./astra-cli proxy stop
./astra-cli proxy remove
```

4. **Restart the container with newly pulled image**
```
./astra-cli proxy quickstart --listen-port 8181 --env-file proxyEnv
```

### Troubleshooting

1. **traces are not captured by astra-proxy-service**

  **Symptoms**

* I have configured the astra-proxy-service upstream proxy address in my Postman/Burp/Curl. I don't see any API endpoint entry in my inventory when I run my postman collection.

  **Cause**

* Potential problem with env file

* astra-traffic-collector is unable to forward the traces to Astra. Refer [here](https://help.getastra.com/en/article/how-to-setup-astra-traffic-collector-in-vm-1dile2q/?bust=1732794298467#3-troubleshooting)

  **Solution**

* Ensure right SENSOR_ID, ALLOWED_HOSTS and OTEL_EXPORTER_ENDPOINT are set in env file.

* Double check if the hostname is registered under **Scope URI for Report** in Target setup page
