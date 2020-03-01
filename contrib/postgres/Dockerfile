FROM postgres
RUN \
  DEBIAN_FRONTEND=noninteractive make-ssl-cert generate-default-snakeoil --force-overwrite \
  && adduser postgres ssl-cert \
  && ln -s /etc/ssl/certs/ssl-cert-snakeoil.pem /var/lib/postgresql/server.crt \
  && ln -s /etc/ssl/private/ssl-cert-snakeoil.key /var/lib/postgresql/server.key
