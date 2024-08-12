# SSL

## This Bash script is to manage SSL certifications.
### It can perform the following tasks:


1. Obtain SSL.


2. Revoke SSL.


3. Force Renew SSL.


## Prerequisites

### Ensure that the `sudo` and `wget` packages are installed on your system:

- Ubuntu & Debian:
```
sudo apt update -q && sudo apt install -y sudo wget
```
- CentOS & Fedora:
```
sudo dnf up -y && sudo dnf install -y sudo wget
```


## Run
#### **Tested on:** Ubuntu 20+, Debian 11+

#### Root Access is Required. If the user is not root, first run:
```
sudo -i
```
#### Then:
```
wget "https://raw.githubusercontent.com/hawshemi/ssl/main/ssl.sh" -O ssl.sh && chmod +x ssl.sh && bash ssl.sh 
```


## Menu Image
![SSL Menu](https://github.com/hawshemi/SSL/assets/16742123/6bc9c28e-32fd-4569-b182-d29f952a5f09)



## Disclaimer
This script is provided as-is, without any warranty or guarantee. You can use it at your own risk.


## License
This script is licensed under the MIT License.

