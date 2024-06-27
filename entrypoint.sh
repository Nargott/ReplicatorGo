#!/bin/sh

set -x
set -e

[ -z "${SIGNAL_BOT_CONFIG_DIR}" ] && echo "SIGNAL_BOT_CONFIG_DIR environmental variable needs to be set! Aborting!" && exit 1;

#usermod -u ${SIGNAL_BOT_UID} signal-bot
#groupmod -g ${SIGNAL_BOT_GID} signal-bot
#
## Fix permissions to ensure backward compatibility
#chown ${SIGNAL_BOT_UID}:${SIGNAL_BOT_GID} -R ${SIGNAL_BOT_CONFIG_DIR}

export HOST_IP=$(hostname -I | awk '{print $1}')

# Start BOT as signal-bot user
#exec setpriv --reuid=${SIGNAL_BOT_UID} --regid=${SIGNAL_BOT_GID} --init-groups /bin/server -cp=${SIGNAL_BOT_CONFIG_DIR}
exec /bin/server -cp=${SIGNAL_BOT_CONFIG_DIR}/config.json


