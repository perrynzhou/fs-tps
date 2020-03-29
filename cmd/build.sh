rm -rf ~/Source/perryn/glusterfs-tps/vendor
go mod vendor
go build -o tps  -mod=vendor
rm -rf ~/Source/perryn/glusterfs-tps/vendor
