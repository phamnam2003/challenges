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
