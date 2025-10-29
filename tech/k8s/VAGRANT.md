# Installation

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

## Uninstall

- As a super user, force remove the Vagrant directories.

```bash
rm -rf /opt/vagrant
rm -f /usr/bin/vagrant
```

- Remove user data: the removal of user data will remove all boxes, plugins, license files, and any stored state that may be used by Vagrant. Removing the user data effectively makes Vagrant think it is a fresh install.
- On Linux and Mac OS platforms, the user data directory location is in the root of your home directory under vagrant.d. Remove the `~/.vagrant.d` directory to delete all the user data. On Windows, this directory is, `C:\Users\YourUsername\.vagrant.d`, where `YourUsername` is the username of your local user.
