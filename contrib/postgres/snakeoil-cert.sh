ln -s /etc/ssl/certs/ssl-cert-snakeoil.pem $PGDATA/server.crt
ln -s /etc/ssl/private/ssl-cert-snakeoil.key $PGDATA/server.key
echo "HERE"
ls $PGDATA
