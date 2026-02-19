init=../init/grd.init
initramfsPath=initramfs
output=initramfs.cpio.gz
sysDir=proc sys tmp dev newRoot

$(output): $(initramfsPath)/init
	@mkdir -p initramfs
	@cd initramfs && mkdir -p $(sysDir)
	@cd initramfs && find . | cpio -ov -H newc | gzip > ../initramfs.cpio.gz

$(initramfsPath)/init: $(init)
	@echo "CP $< -> $@"
	@cp $< $@