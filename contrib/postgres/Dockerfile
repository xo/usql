FROM postgres

COPY snakeoil-cert.sh /docker-entrypoint-initdb.d/

RUN \
  DEBIAN_FRONTEND=noninteractive make-ssl-cert generate-default-snakeoil --force-overwrite \
  && adduser postgres ssl-cert
