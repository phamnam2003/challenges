## Installation

- Install `vagrant` and `virtualbox` with package manager:

```bash
paru -S vagrant
sudo pacman -S virtualbox virtualbox-host-dkms virtualbox-guest-iso
```

- Load module VirtualBox in Kernel:

```bash
sudo modprobe vboxdrv
sudo modprobe vboxnetflt
sudo modprobe vboxnetadp
```

- Run machine with vagrant:

```bash
vagrant up
```

## Usage

- Initialize a new Vagrant environment in the current directory:

```bash
vagrant init <image_name>
# vagrant init ubuntu/jammy64
```

- Start and provision the Vagrant environment:

```bash
vagrant up
```

- ssh to virtual machine:

```bash
vagrant ssh <machine_name>
```

## Destroy

- Trace port forwarding and kill this process:

```bash
sudo lsof -i :<port_number_forwarded>
kill -9 <PID>
```

- Destroy the Vagrant environment:

```bash
vagrant destroy -f 
```

## Uninstall

- As a super user, force remove the Vagrant directories.

```bash
rm -rf /opt/vagrant
rm -f /usr/bin/vagrant
```

- Remove user data: the removal of user data will remove all boxes, plugins, license files, and any stored state that may be used by Vagrant. Removing the user data effectively makes Vagrant think it is a fresh install.
- On Linux and Mac OS platforms, the user data directory location is in the root of your home directory under vagrant.d. Remove the `~/.vagrant.d` directory to delete all the user data. On Windows, this directory is, `C:\Users\YourUsername\.vagrant.d`, where `YourUsername` is the username of your local user.
