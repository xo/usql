FROM apacheignite/ignite

COPY config.xml /opt/ignite
COPY ssl /opt/ignite

CMD $IGNITE_HOME/run.sh /opt/ignite/config.xml
