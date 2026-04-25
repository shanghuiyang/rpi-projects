pushd /home/pi

cp gpstracker.log gpstracker.log.bak
nohup sudo ./gpstracker > gpstracker.log 2>&1 &

nohup ./ip > ip.log 2>&1 &

popd
