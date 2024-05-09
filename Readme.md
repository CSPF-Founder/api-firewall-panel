# API Protector

## Prerequisites

1. Install Vagrant from the official site, https://developer.hashicorp.com/vagrant/downloads. 

- Please refer to this Installation guide if you face any issues during installation. https://developer.hashicorp.com/vagrant/docs/installation

2. Install Virtualbox from the official site, https://www.virtualbox.org/wiki/Downloads

## Minimum Spec

- 4 GB RAM 
- 4 CPU cores
- 10 GB of free disk space

Note: More firewalls/traffic will require more RAM and CPU. Please adjust the Vagrantfile (line number: 58,59) as needed

## Installing VM

Download this repository via 

`git clone https://github.com/CSPF-Founder/api-protector-with-controller-and-panel.git`

Or you can download it as a zip file by clicking on `Code` in the top right and clicking `Download zip`.

`cd` into the folder that is created.

### In Linux:

In the project folder run the below command.

```
chmod +x setupvm.sh

./setupvm.sh
```
During the installation and startup, it will ask you to select the bridge adapter interface as shown below:

> **Which interface should the network bridge to?.**

Please enter the corresponding number for the bridge interface option and press `enter`.

Once the vagrant installation is completed, it will automatically restart in Linux. 

### In Windows:

Go to the project folder on command prompt and then run the below commands.


```
vagrant up
```

During the installation and startup, it will ask you to select the bridge adapter interface as shown below:

> **Which interface should the network bridge to?.**


Please enter the corresponding number for the bridge interface option and press `enter`.

After it has been completed, run the below command to reload the VM manually.

```
vagrant reload
```


## Accessing the Panel

The API Protector Panel is available on this URL: https://localhost:18443. 

```
Note: If you want to change the port, you can change forwardport in the vagrantfile.
```

For information on how to use the panel refer to [Manual.md](Manual.md)

## Further Reading:


- It is highly recommended to change the default password of the user `vagrant` and change the SSH keys. 

- If you want to start the VM after your computer restarts you can give `vargant up` on this folder or start from the virtualbox manager. 

- Once up you can access the VM by giving the command `vagrant ssh apiprotectorvm`

## Contributors

Sabari Selvan

Suriya Prakash
