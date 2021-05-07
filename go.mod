module github.com/f-secure-foundry/armory-boot

go 1.16

require (
	github.com/dsoprea/go-ext4 v0.0.0-20190528173430-c13b09fc0ff8
	github.com/dsoprea/go-logging v0.0.0-20200710184922-b02d349568dd // indirect
	github.com/f-secure-foundry/crucible v0.0.0-20210412160519-6e04f63398d9 // indirect
	github.com/f-secure-foundry/tamago v0.0.0-20210507073204-a2e021e0e2f6
	github.com/flynn/hid v0.0.0-20190502022136-f1b9b6cc019a
	github.com/go-errors/errors v1.2.0 // indirect
	github.com/u-root/u-root v7.0.0+incompatible
	golang.org/x/crypto v0.0.0-20210506145944-38f3c27a63bf
	golang.org/x/net v0.0.0-20210505214959-0714010a04ed // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
)

replace github.com/f-secure-foundry/tamago => /mnt/git/public/tamago
