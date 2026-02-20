
rootfsPath=rootfs
output=disk.img
sysDir=proc tmp dev sys sbin etc
umagCompBin=../umag/grd.umag
umagBin=rootfs/sbin/umag

$(output): $(umagBin)
	@mkdir -p rootfs
	@cd $(rootfsPath) && mkdir -p $(sysDir)
	@mke2fs -F -t ext4 -d rootfs/ disk.img 30M

$(umagBin): $(umagCompBin)
	@cd $(rootfsPath) && mkdir -p sbin
	@cp $< $@