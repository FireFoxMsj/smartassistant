#!/bin/bash

# usage init-efi.sh /mnt/disk/openwrt.img /dev/sda
#image openwrt-21.02.1-x86-64-generic-ext4-combined-efi.img
#build with
#64MB Kernel partition
#384MB Root partition
IMAGE=$1

#target 30GB SSD
#Device         Start      End  Sectors  Size Type
#/dev/sda1       2048   133119   131072   64M EFI System
#/dev/sda2     133120   264191   131072   64M EFI System
#/dev/sda3     264192  4458495  4194304    2G Linux filesystem
#/dev/sda4    4458496  4589567   131072   64M EFI System
#/dev/sda5    4589568  8783871  4194304    2G Linux filesystem
#/dev/sda6    8783872 17172479  8388608    4G Linux swap
#/dev/sda7   17172480 54921216 37748737   18G Linux filesystem
#/dev/sda128       34      511      478  239K BIOS boot

TARGET=$2

echo
echo "Init disk $TARGET from image $IMAGE"
echo

if [ ! -f "$IMAGE" ]; then
    echo "image $IMAGE not found!"
    exit 1
fi

if [ ! -b "$TARGET" ]; then
    echo "disk $TARGET not found or not a block file"
    exit 1
fi

read -p "Is it OK?(y/N):" start
if [ "${start}" == "Y" ] || [ "${start}" == "y" ]; then
    echo "Start Init..."
else 
    echo "Stop!"
    exit 0
fi


sgdisk -o -p "$TARGET"
sgdisk -a 1 -n 128:34:511 -t 128:ef02 -p "$TARGET"
sgdisk -a 2048 -n 1:2048:133119 -t 1:ef00 \
-n 2:133120:264191 -t 2:ef00 \
-n 3:264192:4458495 -t 3:8300 \
-n 4:4458496:4589567 -t 4:ef00 \
-n 5:4589568:8783871 -t 5:8300 \
-n 6:8783872:17172479 -t 6:8200 \
-n 7:17172480:54921216 -t 7:8300 -p "$TARGET"

dd if=$IMAGE of=/dev/sda128 bs=1K skip=17 count=239
dd if=$IMAGE of=/dev/sda1 bs=16K skip=16 count=4096
dd if=$IMAGE of=/dev/sda3 bs=16K skip=4128 count=24576

mkfs.fat /dev/sda2
mkfs.fat /dev/sda4
mkfs.ext4 /dev/sda5 
mkfs.ext4 /dev/sda7
mkswap /dev/sda6 

cd /mnt/
mkdir -p sda1 sda2 sda3 sda4
mount /dev/sda1 sda1
mount /dev/sda2 sda2
mount /dev/sda4 sda4
cp sda1/boot/vmlinuz sda2/
cp sda1/boot/vmlinuz sda4/
e2fsck -f /dev/sda3
resize2fs /dev/sda3
mount /dev/sda3 sda3
 
mkdir -p  data
mount /dev/sda7 data

ROOTFS=$(blkid -o value -s PARTUUID "${TARGET}3")
cat > /mnt/sda1/boot/grub/grub.cfg <<EOF
serial --unit=0 --speed=115200 --word=8 --parity=no --stop=1 --rtscts=off
terminal_input console serial; terminal_output console serial

set default="0"
set timeout="5"
set root='(hd0,gpt2)'

menuentry "SmartAssistant" {
	linux /vmlinuz root=PARTUUID=${ROOTFS} rootwait   console=tty0 console=ttyS0,115200n8 noinitrd
}
menuentry "SmartAssistant (failsafe)" {
	linux /vmlinuz failsafe=true root=PARTUUID=${ROOTFS} rootwait   console=tty0 console=ttyS0,115200n8 noinitrd
}
EOF

echo "Finish!"
echo "You need to edit /etc/config/fstab and make /mnt/data mount shared after reboot"
echo "Using block detect, block mount..."
