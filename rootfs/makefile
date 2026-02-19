
rootfsPath=rootfs
output=disk.img
sysDir=proc tmp dev sys

$(output):
	@mkdir -p rootfs
	@cd $(rootfsPath) && mkdir -p $(sysDir)
	@mke2fs -t ext4 -d rootfs/ disk.img 30M