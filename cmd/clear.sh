ps -ef|grep test|grep -v grep|awk '{print $2}'|xargs kill -9
rm -rf ~/Source/perryn/glusterfs-tps/vendor