FROM rabbitmq:latest

ARG RABBITMQ_USER
ARG RABBITMQ_PASSWORD
ARG CONF_PATH=/etc/rabbitmq/rabbitmq.conf

WORKDIR /

COPY rabbitmq.conf $CONF_PATH
RUN echo "default_user = $RABBITMQ_USER\ndefault_pass = $RABBITMQ_PASSWORD" > $CONF_PATH

CMD ["rabbitmq-server", "$CONF_PATH"]