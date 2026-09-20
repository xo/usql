ARG BASE_IMAGE
FROM $BASE_IMAGE

ARG SCHEMA_URL
ARG TARGET
ARG USER
ADD --chown=$USER $SCHEMA_URL $TARGET/
RUN [ ! -d "$TARGET" ] || chmod -R 777 $TARGET/ || echo "failed to change perms of $TARGET, leaving as $(ls -la $TARGET/)"
