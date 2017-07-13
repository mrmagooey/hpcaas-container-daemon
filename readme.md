# HPCaaS-Container-Daemon

This daemon lives within the HPCaaS container and communicates through its https endpoint. It allows for a properly authenticated external service query or configure the code running in the container, and also provides services internally to the container and the code.

The design philosophy of the container is that it should minimise the amount of changes that the HPC developer should have to make to their code in order to run it in the HPCaaS container. This means that the code should remain unaware that it is running in a container, and things like MPI communication should work seamlessly.

The setup and configuration information below is only for internal developer purposes, the end-user shouldn't need to be manually sending https requests to the container to configure it. The setup should be handled for the end-user either by the HPCaaS local orchestrator in the case of local development, or by a cloud capable HPCaaS system.

## Design

### Communication setup
The daemon will bind to the containers external interface on port 443, and expose an https interface using the self signed certificate. The daemon will also expect all requests to come with an AUTHORIZATION header in them, a variable that also needs to be passed to the container using environment variables.

This information must be supplied to the container using environment variable when the container is first run. Using environment variables, the communication information will be passed as:

    TLS_PUBLIC_CERT=<cert information>
    TLS_PRIVATE_KEY=<key information>
    AUTHORIZATION=<auth password>

The idea is to minimize how much information is passed to the container via the docker run interface, and to prefer the daemon https interface for communication with the container. 

### Container and Daemon setup

Once the container has started, the daemon will be listening for configuration information on its https endpoint. The configuration information is:

1. World Rank (Required): A unique, monotonically increasing integer ID that is given to each container.

1. World Size (Required): The total number of containers in the cluster.

1. Cluster Container IPs (Required): A list of the other container IPs in this cluster. The daemon will add this to the ssh config file as a host. The container will assume that these IP's are reachable from this container.

1. SSH Private and Public Keys (Required): A private and public ssh key-pair. The daemon will add this key-pair to the ssh config file to allow passwordless ssh access between the containers (typically an MPI requirement).

1. Results directory: Where on the container filesystem the daemon will move the results file when the code has finished. This should correspond to a special volume that has been mounted (e.g. host filesytem or network drive).

1. URL: The url that the daemon will upload results to when the code is finished.

### Running Code

When given the `start` command the daemon will run `/hpcaas/code/<hpc code name>`. When a HPCaaS container is created, the HPC code will need to be COPY'ed to this location, as per the container template instructions. This executable does not need to be the executable itself, i.e. it can be a shell script that calls the actual process. However, whatever the executable at  `/hpcaas/code/<hpc code name>` returns will be what the deamon is monitoring. If a non-zero exit code is returned from this executable, the daemon will assume there has been an error and will update the containers code status to `error`.

### Container states

There are several states that the daemon tracks the container as having.

**Container States**

| State   | Description                                           |
|---------|-------------------------------------------------------|
| FSError | The distributed file system has failed                |
| Running | All container services are properly running           |

**Code States**

| State   | Description                                                             |
|---------|-------------------------------------------------------------------------|
| Waiting | The initial state, the daemon is "waiting" to be told to start the code |
| Running | The code is running                                                     |
| Stopped | The code has stopped                                                    |
| Killed  | The code was forcibly killed by the daemon                              |
| Error   | The code has finished with a return code other than 0                   |

**Result States**

| State     | Description                                                                                                |
|-----------|------------------------------------------------------------------------------------------------------------|
| Waiting   | The initial state, the code is still running.                                                              |
| Uploading | The code has finished running, and the container daemon is uploading the results to the specified location |
| Stopped   | The upload has completed successfully                                                                      |
| Error     | There has been an error whilst uploading results                                                           |


## HTTPS Endpoints
### API V1

The json schemas are located in `api/apiv1/schemas`.

*POST /v<api version>/code-parameters*

The user can supply configuration to the container. This will consist of either key value pairs or files. 

Once configuration items in key-value format are received:

1. A json file at `/hpcaas/parameters/parameters.json` will be updated with the new parameter/s
1. A newline separated file at `/hpcaas/parameters/parameters` will be updated with the new parameter/s
1. If the code is not already running, when it is run this parameter will be in its environment variables under the prefix HPCAAS_.

*POST /v1/code-parameter-file*

The user can supply files to the container. These files can be any size or type. Setting the Content-Type to `multipart/form-data`, the deamon  will expect a body containing one or more files. Once received it will be moved to /hpcaas/files/<file-name>, and made world readable and writable.

*POST /v1/daemon-configuration*

Configuration for the daemon. Where code configuration will accept any combination of key and value, the daemon only accepts specific configuration items, and will ignore those that it does not recognise as valid.

*POST /v1/command*

The commands it can receive are:

| Command | Effect                                                                                                |
|---------|-------------------------------------------------------------------------------------------------------|
| Start   | Will run the executable file. Requires code state to be "Waiting", otherwise command will be ignored. |
| Kill    | Will forcibly kill the code process. Will put the code state to "Killed".                             |

It will expect a JSON schema as follows:

*GET /v1/query*

To which it will provide information about the running code and container:

| Metric          | Description                                  |
|-----------------|----------------------------------------------|
| containerStatus | The container status                         |
| codeStatus      | The code status                              |
| resultStatus    | The result status                            |
| percentComplete | How far through the computations the code is |
|                 |                                              |

## Performance Impact
The daemon has a minimal memory impact, and effectively consists of a set of event listeners which trigger infrequently whilst the code is running and perform minimal work when they do trigger.

