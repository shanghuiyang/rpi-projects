
pushd /home/pi

cp xcar.log xcar.log.bak
nohup sudo ./xcar > xcar.log 2>&1 &

popd

sudo motion
