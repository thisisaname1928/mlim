buildTools=grd.buildTools

run: $(buildTools)
	@./$<
	@qemu-system-x86_64 -kernel bzImage -initrd initramfs/initramfs.cpio.gz -drive format=raw,file=rootfs/disk.img -append "quiet root=/dev/sda"

$(buildTools):
	@echo "build buildTools..."
	@cd buildTools && go build .
	@cp buildTools/$(buildTools) $@