
cp initd.sh /etc/init.d/incus

if [ ! -d "/etc/incus" ]; then
    mkdir /etc/incus
fi
cp default.conf /etc/incus/incus.conf
cp incus /usr/sbin/incus

touch /var/log/incus.log


