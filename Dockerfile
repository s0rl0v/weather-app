FROM alpine

ARG uid=1000
ARG gid=3000

WORKDIR /app

# https://stackoverflow.com/a/49955098
# https://github.com/brgl/busybox/blob/master/loginutils/adduser.c
# https://github.com/brgl/busybox/blob/master/loginutils/addgroup.c

RUN addgroup --gid ${gid} app-group && \
    adduser --uid ${uid} --ingroup app-group --disabled-password app

USER ${uid}:${gid}

COPY --chown=${uid}:${gid} server .

# Make entrypoint executable for self-contained apps
RUN chmod +x "/app/server"

ENTRYPOINT /app/server
