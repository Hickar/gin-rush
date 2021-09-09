FROM bitnami/redis:latest

ARG REDIS_PASSWORD=V3nomous
ARG CONF_PATH=/usr/local/etc/redis/redis.conf

WORKDIR /

COPY redis.conf /usr/local/etc/redis/redis.conf
RUN echo "requirepass $REDIS_PASSWORD" > /usr/local/etc/redis/redis.conf

CMD ["redis-server", "/usr/local/etc/redis/redis.conf"]