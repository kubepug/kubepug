# Database

The database is a simple generated json file containing the APIs marked as deprecated and deleted.

It can be downloaded from [here](https://kubepug.xyz/data/data.json).

## Generating my own database

In case you want to generate your own data.json file, follow the steps below. 
We use a [container image](https://github.com/rikatz/kubepug/blob/main/generator/Dockerfile) so the whole step can be reproduced locally.
 
1. Clone/Download this repository, and build the container on `generator/` directory
  ```console
  git clone https://github.com/rikatz/kubepug
  docker build -t generator -f generator/Dockerfile generator
  ```
1. Generate the data.json
  ```console
  docker run generator > data.json
  ```
  Generator uses the latest stable Kubernetes API version, if you want the latest dev version you should run as:
  ```
  docker run -e VERSION=master generator > data.json
  ```
1. Securely move the json file to your Air-Gapped environment, to the folder of your choosing. This folder will be used by `kubepug`.
1. Execute `kubepug` with the option `database`, like this
  ```console
  kubepug --k8s-version=v1.22 --database=location/of/your/data.json
  ```
